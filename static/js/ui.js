/**
 * UI 渲染模块
 * 负责页面显示和更新
 */

import { state } from './config.js';

/**
 * 显示记录列表
 * @param {Array} records - 记录数组
 */
export function displayRecords(records) {
    const tbody = document.getElementById('tableBody');

    if (!records || records.length === 0) {
        tbody.innerHTML = '<tr><td colspan="11" class="no-data">当前月份暂无数据</td></tr>';
        return;
    }

    tbody.innerHTML = records.map(record => {
        // 格式化额外费用显示
        let extraFeesDisplay = '-';
        if (record.extraFees && record.extraFees.length > 0) {
            const extraTotal = record.extraFees.reduce((sum, fee) => sum + fee.amount, 0);
            const feeNames = record.extraFees.map(fee => fee.name).join(', ');
            extraFeesDisplay = `<span title="${feeNames}">¥${extraTotal.toFixed(2)} (${record.extraFees.length}项)</span>`;
        }

        return `
        <tr>
            <td><input type="checkbox" class="record-checkbox" value="${record.id}"></td>
            <td><strong>${record.roomNumber}</strong></td>
            <td>${record.billingMonth}</td>
            <td>${record.waterUsage.toFixed(2)}</td>
            <td>¥${record.totalWaterCost.toFixed(2)}</td>
            <td>${record.electricUsage.toFixed(2)}</td>
            <td>¥${record.totalElectricCost.toFixed(2)}</td>
            <td>¥${record.managementFee.toFixed(2)}</td>
            <td>${extraFeesDisplay}</td>
            <td><strong>¥${record.totalCost.toFixed(2)}</strong></td>
            <td>
                <div class="action-buttons">
                    <button class="btn btn-warning" onclick="window.billingApp.editRecord(${record.id})">编辑</button>
                    <button class="btn btn-danger" onclick="window.billingApp.deleteRecord(${record.id})">删除</button>
                    <button class="btn btn-report" onclick="window.billingApp.generateSingleCard(${record.id})">📊</button>
                </div>
            </td>
        </tr>
    `}).join('');

    // 为所有复选框添加事件监听
    setTimeout(() => {
        document.querySelectorAll('.record-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', updateSelectedStatistics);
        });
        // 初始化选中统计
        updateSelectedStatistics();
    }, 0);
}

/**
 * 更新统计信息
 * @param {Array} records - 记录数组
 */
export function updateStatistics(records) {
    const total = records.length;
    const totalCost = records.reduce((sum, r) => sum + r.totalCost, 0);
    const totalWater = records.reduce((sum, r) => sum + r.totalWaterCost, 0);
    const totalElectric = records.reduce((sum, r) => sum + r.totalElectricCost, 0);

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
        // 没有选中任何记录，隐藏统计面板
        selectedPanel.style.display = 'none';
        return;
    }

    // 显示统计面板
    selectedPanel.style.display = 'block';

    // 获取选中的记录
    const selectedRecords = state.allRecords.filter(record => selectedIds.includes(record.id));

    // 计算统计数据
    const count = selectedRecords.length;
    const totalWaterCost = selectedRecords.reduce((sum, r) => sum + r.totalWaterCost, 0);
    const totalElectricCost = selectedRecords.reduce((sum, r) => sum + r.totalElectricCost, 0);
    const totalManagementFee = selectedRecords.reduce((sum, r) => sum + r.managementFee, 0);

    // 计算额外费用总和
    let totalExtraFee = 0;
    selectedRecords.forEach(record => {
        if (record.extraFees && record.extraFees.length > 0) {
            totalExtraFee += record.extraFees.reduce((sum, fee) => sum + fee.amount, 0);
        }
    });

    const totalCost = selectedRecords.reduce((sum, r) => sum + r.totalCost, 0);

    // 更新显示
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

    // 获取所有唯一的房号
    const rooms = [...new Set(state.allRecords.map(r => r.roomNumber))].sort();

    // 保存当前选中的值
    const currentValue = roomFilter.value;

    // 清空并重新填充
    roomFilter.innerHTML = '<option value="">全部房号</option>';
    rooms.forEach(room => {
        const option = document.createElement('option');
        option.value = room;
        option.textContent = room;
        roomFilter.appendChild(option);
    });

    // 恢复选中的值
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
    // 如果从悬浮栏调用，强制全选
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
    tbody.innerHTML = '<tr><td colspan="11" class="no-data">请先选择月份查看账单数据</td></tr>';
}