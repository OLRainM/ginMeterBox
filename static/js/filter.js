/**
 * 筛选和排序模块
 */

import { state } from './config.js';
import { displayRecords, updateStatistics } from './ui.js';
import { fetchRecords } from './api.js';

/**
 * 应用组合筛选（月份 + 房号）
 */
export function applyFilters() {
    const month = document.getElementById('monthFilter').value;
    const room = document.getElementById('roomFilter').value;
    
    let filtered = state.allRecords;
    
    // 按月份筛选
    if (month) {
        filtered = filtered.filter(record => record.billingMonth === month);
    }
    
    // 按房号筛选
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
    displayRecords(state.allRecords);
    updateStatistics(state.allRecords);
}

/**
 * 按房号排序
 * @param {string} order - 排序方式 (asc/desc)
 * @param {Function} onSuccess - 成功回调
 */
export async function sortByRoom(order, onSuccess) {
    state.currentSortOrder = order;
    const records = await fetchRecords(order);
    state.allRecords = records;
    
    displayRecords(records);
    updateStatistics(records);
    
    if (onSuccess) onSuccess(`已按房号${order === 'asc' ? '升序' : '降序'}排序`);
}
