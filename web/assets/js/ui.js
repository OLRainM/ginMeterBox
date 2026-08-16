/**
 * UI 渲染模块
 * 负责页面显示和更新
 */

import { state } from './config.js';

function createCell(content) {
    const cell = document.createElement('td');
    if (content instanceof Node) {
        cell.appendChild(content);
    } else {
        cell.textContent = content;
    }
    return cell;
}

function createButton(className, text, onClick) {
    const button = document.createElement('button');
    button.className = className;
    button.type = 'button';
    button.textContent = text;
    button.addEventListener('click', onClick);
    return button;
}

/**
 * 显示记录列表
 * @param {Array} records - 记录数组
 */
export function displayRecords(records) {
    const tbody = document.getElementById('tableBody');
    const householdRecords = (records || []).filter(record => record.roomNumber !== '总表');

    if (householdRecords.length === 0) {
        const row = document.createElement('tr');
        const cell = document.createElement('td');
        cell.colSpan = 11;
        cell.className = 'no-data';
        cell.textContent = '当前月份暂无数据';
        row.appendChild(cell);
        tbody.replaceChildren(row);
        return;
    }

    const rows = householdRecords.map(record => {
        const row = document.createElement('tr');
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'record-checkbox';
        checkbox.value = record.id;
        checkbox.addEventListener('change', updateSelectedStatistics);
        row.appendChild(createCell(checkbox));

        const roomNumber = document.createElement('strong');
        roomNumber.textContent = record.roomNumber;
        row.appendChild(createCell(roomNumber));
        row.appendChild(createCell(record.billingMonth));

        const rawWater = (record.currentWater - record.previousWater).toFixed(1);
        const waterUsage = document.createDocumentFragment();
        const rawWaterSpan = document.createElement('span');
        rawWaterSpan.className = 'raw-val';
        rawWaterSpan.textContent = rawWater;
        const waterUsageSpan = document.createElement('span');
        waterUsageSpan.className = 'adj-val';
        waterUsageSpan.textContent = record.waterUsage.toFixed(1);
        waterUsage.append(rawWaterSpan, waterUsageSpan);
        row.appendChild(createCell(waterUsage));
        row.appendChild(createCell(`¥${record.totalWaterCost.toFixed(2)}`));

        const rawElectric = (record.currentElectric - record.previousElectric).toFixed(1);
        const electricUsage = document.createDocumentFragment();
        const rawElectricSpan = document.createElement('span');
        rawElectricSpan.className = 'raw-val';
        rawElectricSpan.textContent = rawElectric;
        const electricUsageSpan = document.createElement('span');
        electricUsageSpan.className = 'adj-val';
        electricUsageSpan.textContent = record.electricUsage.toFixed(1);
        electricUsage.append(rawElectricSpan, electricUsageSpan);
        row.appendChild(createCell(electricUsage));
        row.appendChild(createCell(`¥${record.totalElectricCost.toFixed(2)}`));
        row.appendChild(createCell(`¥${record.managementFee.toFixed(2)}`));

        const extraFeesCell = document.createElement('td');
        if (record.extraFees && record.extraFees.length > 0) {
            const extraTotal = record.extraFees.reduce((sum, fee) => sum + fee.amount, 0);
            const extraFeesDisplay = document.createElement('span');
            extraFeesDisplay.title = record.extraFees.map(fee => fee.name).join(', ');
            extraFeesDisplay.textContent = `¥${extraTotal.toFixed(2)} (${record.extraFees.length}项)`;
            extraFeesCell.appendChild(extraFeesDisplay);
        } else {
            extraFeesCell.textContent = '-';
        }
        row.appendChild(extraFeesCell);

        const totalCost = document.createElement('strong');
        totalCost.textContent = `¥${record.totalCost.toFixed(2)}`;
        row.appendChild(createCell(totalCost));

        const actionCell = document.createElement('td');
        const actionButtons = document.createElement('div');
        actionButtons.className = 'action-buttons';
        actionButtons.append(
            createButton('btn btn-warning', '编辑', () => window.billingApp.editRecord(record.id)),
            createButton('btn btn-danger', '删除', () => window.billingApp.deleteRecord(record.id)),
            createButton('btn btn-report', '📊', () => window.billingApp.generateSingleCard(record.id))
        );
        actionCell.appendChild(actionButtons);
        row.appendChild(actionCell);

        return row;
    });

    tbody.replaceChildren(...rows);
    updateSelectedStatistics();
}

/**
 * 更新统计信息
 * @param {Array} records - 记录数组
 */
export function updateStatistics(records) {
    // “总表”是独立计量项，不计入27户住户的总数和费用汇总。
    const householdRecords = records.filter(record => record.roomNumber !== '总表');
    const total = householdRecords.length;
    const totalCost = householdRecords.reduce((sum, r) => sum + r.totalCost, 0);
    const totalWater = householdRecords.reduce((sum, r) => sum + r.totalWaterCost, 0);
    const totalElectric = householdRecords.reduce((sum, r) => sum + r.totalElectricCost, 0);

    document.getElementById('totalRecords').textContent = total;
    document.getElementById('totalCost').textContent = `¥${totalCost.toFixed(2)}`;
    document.getElementById('totalWater').textContent = `¥${totalWater.toFixed(2)}`;
    document.getElementById('totalElectric').textContent = `¥${totalElectric.toFixed(2)}`;
}

/**
 * 更新选中记录统计
 */
export function updateSelectedStatistics() {
    const selectedIds = Array.from(document.querySelectorAll('.record-checkbox:checked'))
        .map(cb => parseInt(cb.value));
    const selectedPanel = document.getElementById('selectedStatistics');

    if (selectedIds.length === 0) {
        selectedPanel.style.display = 'none';
        return;
    }

    selectedPanel.style.display = 'block';
    const selectedRecords = state.allRecords.filter(record => record.roomNumber !== '总表' && selectedIds.includes(record.id));
    const count = selectedRecords.length;
    const totalWaterCost = selectedRecords.reduce((sum, r) => sum + r.totalWaterCost, 0);
    const totalElectricCost = selectedRecords.reduce((sum, r) => sum + r.totalElectricCost, 0);
    const totalManagementFee = selectedRecords.reduce((sum, r) => sum + r.managementFee, 0);

    let totalExtraFee = 0;
    selectedRecords.forEach(record => {
        if (record.extraFees && record.extraFees.length > 0) {
            totalExtraFee += record.extraFees.reduce((sum, fee) => sum + fee.amount, 0);
        }
    });

    const totalCost = selectedRecords.reduce((sum, r) => sum + r.totalCost, 0);

    document.getElementById('selectedCount').textContent = count;
    document.getElementById('selectedWaterCost').textContent = `¥${totalWaterCost.toFixed(2)}`;
    document.getElementById('selectedElectricCost').textContent = `¥${totalElectricCost.toFixed(2)}`;
    document.getElementById('selectedManagementFee').textContent = `¥${totalManagementFee.toFixed(2)}`;
    document.getElementById('selectedExtraFee').textContent = `¥${totalExtraFee.toFixed(2)}`;
    document.getElementById('selectedTotalCost').textContent = `¥${totalCost.toFixed(2)}`;
}

/**
 * 填充房号筛选下拉框
 */
export function populateRoomFilter() {
    const roomFilter = document.getElementById('roomFilter');
    if (!roomFilter) return;

    const rooms = [...new Set(state.allRecords.filter(r => r.roomNumber !== '总表').map(r => r.roomNumber))].sort();
    const currentValue = roomFilter.value;
    const defaultOption = document.createElement('option');
    defaultOption.value = '';
    defaultOption.textContent = '全部房号';
    roomFilter.replaceChildren(defaultOption);

    rooms.forEach(room => {
        const option = document.createElement('option');
        option.value = room;
        option.textContent = room;
        roomFilter.appendChild(option);
    });

    if (currentValue && rooms.includes(currentValue)) {
        roomFilter.value = currentValue;
    }
}

/**
 * 全选/取消全选
 */
export function toggleSelectAll() {
    const selectAll = document.getElementById('selectAll');
    const checkboxes = document.querySelectorAll('.record-checkbox');
    if (!selectAll.checked) {
        selectAll.checked = true;
    }
    checkboxes.forEach(cb => cb.checked = selectAll.checked);
    updateSelectedStatistics();
}

/**
 * 全选所有可见记录
 */
export function selectAllRecords() {
    document.getElementById('selectAll').checked = true;
    document.querySelectorAll('.record-checkbox').forEach(cb => cb.checked = true);
    updateSelectedStatistics();
}

/**
 * 清除所有选择
 */
export function clearSelectionUI() {
    document.getElementById('selectAll').checked = false;
    document.querySelectorAll('.record-checkbox').forEach(cb => cb.checked = false);
    updateSelectedStatistics();
}

/**
 * 显示未选择月份的空状态
 */
export function showEmptyMonth() {
    const tbody = document.getElementById('tableBody');
    const row = document.createElement('tr');
    const cell = document.createElement('td');
    cell.colSpan = 11;
    cell.className = 'no-data';
    cell.textContent = '请先选择月份查看账单数据';
    row.appendChild(cell);
    tbody.replaceChildren(row);
}
