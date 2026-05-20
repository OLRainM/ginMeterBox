/**
 * 筛选和排序模块
 */

import { state } from './config.js';
import { displayRecords, updateStatistics, showEmptyMonth } from './ui.js';
import { fetchRecords, fetchTotalMeter } from './api.js';

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
        hideDiffPanel();
        return;
    }

    let filtered = state.allRecords.filter(record => record.billingMonth === month);

    if (room) {
        filtered = filtered.filter(record => record.roomNumber === room);
    }

    displayRecords(filtered);
    updateStatistics(filtered);

    // 计算水电差（使用当月所有用户数据，不受房号筛选影响）
    const allMonthRecords = state.allRecords.filter(r => r.billingMonth === month);
    updateDiffPanel(month, allMonthRecords);
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
    hideDiffPanel();
}

/**
 * 按房号排序
 */
export async function sortByRoom(order, onSuccess) {
    state.currentSortOrder = order;
    const records = await fetchRecords(order);
    state.allRecords = records;
    applyFilters();
    if (onSuccess) onSuccess(`已按房号${order === 'asc' ? '升序' : '降序'}排序`);
}

/**
 * 计算并更新水电差面板
 */
async function updateDiffPanel(month, records) {
    const panel = document.getElementById('diffPanel');

    // 获取当月和上月总表数据
    const currentMeter = await fetchTotalMeter(month);
    // 计算上月月份
    const [y, m] = month.split('-').map(Number);
    const prevMonth = m === 1 ? `${y-1}-12` : `${y}-${String(m-1).padStart(2,'0')}`;
    const prevMeter = await fetchTotalMeter(prevMonth);

    if (!currentMeter || !prevMeter) {
        panel.style.display = 'none';
        return;
    }

    // 总表用量 = 当月读数 - 上月读数
    const totalWaterUsage = currentMeter.waterReading - prevMeter.waterReading;
    const totalElectricUsage = currentMeter.electricReading - prevMeter.electricReading;

    // 各户用量之和（表显用量，不含补差）
    const sumUserWater = records.reduce((s, r) => s + (r.currentWater - r.previousWater), 0);
    const sumUserElectric = records.reduce((s, r) => s + (r.currentElectric - r.previousElectric), 0);

    // 差值
    const waterDiff = totalWaterUsage - sumUserWater;
    const electricDiff = totalElectricUsage - sumUserElectric;

    // 补差建议（÷27）
    const waterSuggest = waterDiff / 27;
    const electricSuggest = electricDiff / 27;

    // 更新面板
    document.getElementById('totalMeterWater').textContent = totalWaterUsage.toFixed(1);
    document.getElementById('sumUserWater').textContent = sumUserWater.toFixed(1);
    document.getElementById('waterDiff').textContent = waterDiff.toFixed(1);
    document.getElementById('waterSuggest').textContent = waterSuggest.toFixed(2);
    document.getElementById('totalMeterElectric').textContent = totalElectricUsage.toFixed(1);
    document.getElementById('sumUserElectric').textContent = sumUserElectric.toFixed(1);
    document.getElementById('electricDiff').textContent = electricDiff.toFixed(1);
    document.getElementById('electricSuggest').textContent = electricSuggest.toFixed(2);

    panel.style.display = 'block';
}

function hideDiffPanel() {
    document.getElementById('diffPanel').style.display = 'none';
}
