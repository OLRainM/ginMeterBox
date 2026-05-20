/**
 * 筛选和排序模块
 */

import { state } from './config.js';
import { displayRecords, updateStatistics, showEmptyMonth } from './ui.js';
import { fetchRecords } from './api.js';

/**
 * 应用组合筛选（月份 + 房号）
 */
export function applyFilters() {
    const month = document.getElementById('monthFilter').value;
    const room = document.getElementById('roomFilter').value;

    state.currentMonth = month;

    if (!month) {
        showEmptyMonth();
        updateStatistics([]);
        return;
    }

    let filtered = state.allRecords.filter(record => record.billingMonth === month);

    if (room) {
        filtered = filtered.filter(record => record.roomNumber === room);
    }

    displayRecords(filtered);
    updateStatistics(filtered);
}

/**
 * 清除筛选
 */
export function clearFilter() {
    document.getElementById('monthFilter').value = '';
    document.getElementById('roomFilter').value = '';
    state.currentMonth = '';
    showEmptyMonth();
    updateStatistics([]);
}

/**
 * 按房号排序
 */
export async function sortByRoom(order, onSuccess) {
    state.currentSortOrder = order;
    const records = await fetchRecords(order);
    state.allRecords = records;

    // 重新应用当前月份筛选
    applyFilters();

    if (onSuccess) onSuccess(`已按房号${order === 'asc' ? '升序' : '降序'}排序`);
}
