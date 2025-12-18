/**
 * 智能水表匹配模块
 */

import { showNotification, getSelectedIds } from './utils.js';
import { state } from './config.js';
import { smartWaterMatch } from './api.js';
import { clearSelectionUI } from './ui.js';

/**
 * 显示智能匹配模态框
 */
export function showSmartMatchModal() {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要匹配的用户', 'error');
        return;
    }
    
    // 检查用户数量限制
    if (ids.length > 10) {
        showNotification('为保证性能，单次匹配用户数量不能超过10个，请分批处理', 'error');
        return;
    }
    
    // 获取选中的记录
    const selectedRecords = state.allRecords.filter(r => ids.includes(r.id));
    
    // 显示选中的用户信息
    displaySelectedUsers(selectedRecords);
    
    // 清空输入框
    document.getElementById('waterReadingsInput').value = '';
    document.getElementById('matchPreview').innerHTML = '';
    
    // 显示模态框
    document.getElementById('smartMatchModal').style.display = 'block';
}

/**
 * 关闭智能匹配模态框
 */
export function closeSmartMatchModal() {
    document.getElementById('smartMatchModal').style.display = 'none';
}

/**
 * 显示选中的用户信息
 */
function displaySelectedUsers(records) {
    const container = document.getElementById('selectedUsersList');
    
    const html = `
        <div class="selected-users-info">
            <p><strong>已选择 ${records.length} 个用户：</strong></p>
            <div class="users-grid">
                ${records.map(r => `
                    <div class="user-card">
                        <div class="user-room">${r.roomNumber}</div>
                        <div class="user-detail">上月水表: ${r.previousWater}</div>
                        <div class="user-detail">补差: ${r.waterAdjustment || 0}</div>
                    </div>
                `).join('')}
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

/**
 * 解析水表读数输入
 */
function parseWaterReadings(input) {
    // 支持多种分隔符：空格、逗号、换行
    const readings = input
        .split(/[\s,\n]+/)
        .map(s => s.trim())
        .filter(s => s.length > 0)
        .map(s => parseFloat(s))
        .filter(n => !isNaN(n));
    
    return readings;
}

/**
 * 预览匹配结果
 */
export function previewMatch() {
    const ids = getSelectedIds();
    const input = document.getElementById('waterReadingsInput').value.trim();
    
    if (!input) {
        showNotification('请输入水表读数', 'error');
        return;
    }
    
    const readings = parseWaterReadings(input);
    
    if (readings.length === 0) {
        showNotification('未识别到有效的水表读数', 'error');
        return;
    }
    
    if (readings.length !== ids.length) {
        showNotification(`读数数量(${readings.length})与选中用户数量(${ids.length})不匹配`, 'error');
        return;
    }
    
    // 获取选中的记录
    const selectedRecords = state.allRecords.filter(r => ids.includes(r.id));
    
    // 本地模拟匹配（用于预览）
    const matches = simulateMatch(selectedRecords, readings);
    
    // 显示预览
    displayMatchPreview(matches);
}

/**
 * 本地模拟匹配（用于预览）
 */
function simulateMatch(records, readings) {
    const n = records.length;
    let bestMatches = [];
    let minTotalUsage = Infinity;
    
    // 生成所有排列
    const permutations = generatePermutations(readings);
    
    for (const perm of permutations) {
        let totalUsage = 0;
        const currentMatches = [];
        
        for (let i = 0; i < n; i++) {
            const usage = perm[i] - records[i].previousWater + (records[i].waterAdjustment || 0);
            totalUsage += usage;
            currentMatches.push({
                record: records[i],
                waterReading: perm[i],
                waterUsage: usage
            });
        }
        
        if (totalUsage < minTotalUsage) {
            minTotalUsage = totalUsage;
            bestMatches = currentMatches;
        }
    }
    
    return { matches: bestMatches, totalUsage: minTotalUsage };
}

/**
 * 生成排列组合（使用回溯算法）
 */
function generatePermutations(arr) {
    const result = [];
    const n = arr.length;
    
    // 边界情况
    if (n === 0) return result;
    if (n === 1) return [[arr[0]]];
    
    // 回溯算法
    function backtrack(current, start) {
        if (start === n) {
            // 找到完整排列，复制并添加
            result.push([...current]);
            return;
        }
        
        for (let i = start; i < n; i++) {
            // 交换
            [current[start], current[i]] = [current[i], current[start]];
            // 递归
            backtrack(current, start + 1);
            // 回溯
            [current[start], current[i]] = [current[i], current[start]];
        }
    }
    
    // 创建工作数组
    const working = [...arr];
    backtrack(working, 0);
    
    return result;
}

/**
 * 显示匹配预览
 */
function displayMatchPreview(result) {
    const container = document.getElementById('matchPreview');
    
    const html = `
        <div class="match-preview">
            <h4>🎯 最优匹配方案（总用水量最小）</h4>
            <div class="total-usage">
                <strong>总用水量：</strong>${result.totalUsage.toFixed(2)} 吨
            </div>
            <div class="match-results">
                ${result.matches.map(m => `
                    <div class="match-item">
                        <div class="match-room">${m.record.roomNumber}</div>
                        <div class="match-details">
                            <div>上月: ${m.record.previousWater}</div>
                            <div class="match-arrow">→</div>
                            <div class="match-current">本月: ${m.waterReading}</div>
                            <div class="match-usage">用量: ${m.waterUsage.toFixed(2)} 吨</div>
                        </div>
                    </div>
                `).join('')}
            </div>
            <div class="match-actions">
                <button onclick="billingApp.executeSmartMatch()" class="btn btn-primary">
                    ✅ 确认并应用
                </button>
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

/**
 * 执行智能匹配
 */
export async function executeSmartMatch(onSuccess) {
    const ids = getSelectedIds();
    const input = document.getElementById('waterReadingsInput').value.trim();
    
    if (!input) {
        showNotification('请输入水表读数', 'error');
        return;
    }
    
    const readings = parseWaterReadings(input);
    
    if (readings.length === 0) {
        showNotification('未识别到有效的水表读数', 'error');
        return;
    }
    
    if (readings.length !== ids.length) {
        showNotification(`读数数量(${readings.length})与选中用户数量(${ids.length})不匹配`, 'error');
        return;
    }
    
    const result = await smartWaterMatch(ids, readings);
    
    if (result) {
        closeSmartMatchModal();
        clearSelectionUI();
        if (onSuccess) onSuccess();
    }
}
