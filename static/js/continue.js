/**
 * 自动延续模块
 */

import { state } from './config.js';
import { showNotification } from './utils.js';
import { fetchLatestRecord, continueRecord, batchContinueRecords } from './api.js';

function createPreviewRow(label, value, strong = false) {
    const row = document.createElement('tr');
    const labelCell = document.createElement('td');
    labelCell.textContent = label;
    const valueCell = document.createElement('td');
    const valueElement = strong ? document.createElement('strong') : document.createElement('span');
    valueElement.textContent = value;
    valueCell.appendChild(valueElement);
    row.append(labelCell, valueCell);
    return row;
}

/**
 * 显示自动延续表单
 */
export function showContinueForm() {
    document.getElementById('continueModal').style.display = 'block';
    document.getElementById('previousInfo').style.display = 'none';

    const now = new Date();
    const nextMonth = new Date(now.getFullYear(), now.getMonth() + 1, 1);
    const monthStr = nextMonth.toISOString().slice(0, 7);
    document.getElementById('continueNewMonth').value = monthStr;

    document.querySelector('input[name="continueMode"][value="single"]').checked = true;
    toggleContinueMode();
}

/**
 * 关闭自动延续模态框
 */
export function closeContinueModal() {
    document.getElementById('continueModal').style.display = 'none';
    document.getElementById('continueRoomNumber').value = '';
    document.getElementById('previousInfo').style.display = 'none';
}

/**
 * 切换自动延续模式（单户/批量）
 */
export function toggleContinueMode() {
    const mode = document.querySelector('input[name="continueMode"]:checked').value;
    const singleSection = document.getElementById('singleContinueSection');
    const batchSection = document.getElementById('batchContinueSection');

    if (mode === 'single') {
        singleSection.style.display = 'block';
        batchSection.style.display = 'none';
    } else {
        singleSection.style.display = 'none';
        batchSection.style.display = 'block';
        populateRoomSelectionList();
    }
}

/**
 * 填充房号选择列表
 */
export function populateRoomSelectionList() {
    const container = document.getElementById('roomSelectionList');
    if (!container) return;

    const rooms = [...new Set(state.allRecords.map(r => r.roomNumber))].sort();
    const labels = rooms.map(room => {
        const label = document.createElement('label');
        label.className = 'room-checkbox-label';
        label.style.cssText = 'display: block; padding: 5px 10px; cursor: pointer;';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'room-checkbox';
        checkbox.value = room;
        checkbox.addEventListener('change', updateSelectedRoomsCount);

        label.append(checkbox, document.createTextNode(` ${room}`));
        return label;
    });

    container.replaceChildren(...labels);
    updateSelectedRoomsCount();
}

/**
 * 全选房号
 */
export function selectAllRooms() {
    document.querySelectorAll('.room-checkbox').forEach(cb => cb.checked = true);
    updateSelectedRoomsCount();
}

/**
 * 取消全选房号
 */
export function deselectAllRooms() {
    document.querySelectorAll('.room-checkbox').forEach(cb => cb.checked = false);
    updateSelectedRoomsCount();
}

/**
 * 更新已选择房号数量
 */
export function updateSelectedRoomsCount() {
    const count = document.querySelectorAll('.room-checkbox:checked').length;
    const countSpan = document.getElementById('selectedRoomsCount');
    if (countSpan) {
        countSpan.textContent = `已选择: ${count}`;
    }
}

/**
 * 预览上月数据
 */
export async function previewContinue() {
    const roomNumber = document.getElementById('continueRoomNumber').value.trim();

    if (!roomNumber) {
        showNotification('请输入住户编号', 'error');
        return;
    }

    const record = await fetchLatestRecord(roomNumber);

    if (record) {
        const table = document.createElement('table');
        table.append(
            createPreviewRow('住户编号:', record.roomNumber, true),
            createPreviewRow('上月缴费月份:', record.billingMonth),
            createPreviewRow('当前水表读数:', `${record.currentWater.toFixed(2)} 吨`, true),
            createPreviewRow('当前电表读数:', `${record.currentElectric.toFixed(2)} 度`, true),
            createPreviewRow('管理费:', `¥${record.managementFee.toFixed(2)}`),
            createPreviewRow('水单价:', `¥${record.waterPrice.toFixed(2)}/吨`),
            createPreviewRow('电单价:', `¥${record.electricPrice.toFixed(2)}/度`)
        );

        const message = document.createElement('p');
        message.style.cssText = 'margin-top: 15px; color: #28a745; font-weight: 600;';
        message.textContent = '✅ 新记录将使用这些读数作为上月读数，并初始化本月读数为相同值';

        document.getElementById('previousData').replaceChildren(table, message);
        document.getElementById('previousInfo').style.display = 'block';
    }
}

/**
 * 执行自动延续
 * @param {Function} onSuccess - 成功回调
 */
export async function executeContinue(onSuccess) {
    const mode = document.querySelector('input[name="continueMode"]:checked').value;
    const newMonth = document.getElementById('continueNewMonth').value;

    if (!newMonth) {
        showNotification('请选择新月份', 'error');
        return;
    }

    let success = false;

    if (mode === 'single') {
        const roomNumber = document.getElementById('continueRoomNumber').value.trim();

        if (!roomNumber) {
            showNotification('请输入住户编号', 'error');
            return;
        }

        success = await continueRecord(roomNumber, newMonth);
    } else {
        const selectedRooms = Array.from(document.querySelectorAll('.room-checkbox:checked'))
            .map(cb => cb.value);

        if (selectedRooms.length === 0) {
            showNotification('请至少选择一个住户', 'error');
            return;
        }

        const result = await batchContinueRecords(selectedRooms, newMonth);
        success = result !== null;
    }

    if (success) {
        closeContinueModal();
        if (onSuccess) onSuccess();
    }
}
