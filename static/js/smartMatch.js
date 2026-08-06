/**
 * 智能水表匹配模块
 */

import { showNotification, getSelectedIds } from './utils.js';
import { state } from './config.js';
import { smartWaterMatch } from './api.js';
import { clearSelectionUI } from './ui.js';

function createElement(tagName, className, text) {
    const element = document.createElement(tagName);
    if (className) element.className = className;
    if (text !== undefined) element.textContent = text;
    return element;
}

/**
 * 显示智能匹配模态框
 */
export function showSmartMatchModal() {
    const ids = getSelectedIds();

    if (ids.length === 0) {
        showNotification('请先选择要匹配的用户', 'error');
        return;
    }

    if (ids.length > 10) {
        showNotification('为保证性能，单次匹配用户数量不能超过10个，请分批处理', 'error');
        return;
    }

    const selectedRecords = state.allRecords.filter(r => ids.includes(r.id));
    displaySelectedUsers(selectedRecords);
    document.getElementById('waterReadingsInput').value = '';
    document.getElementById('matchPreview').replaceChildren();
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
    const info = createElement('div', 'selected-users-info');
    const heading = document.createElement('p');
    const headingText = document.createElement('strong');
    headingText.textContent = `已选择 ${records.length} 个用户：`;
    heading.appendChild(headingText);

    const usersGrid = createElement('div', 'users-grid');
    records.forEach(record => {
        const card = createElement('div', 'user-card');
        card.append(
            createElement('div', 'user-room', record.roomNumber),
            createElement('div', 'user-detail', `上月水表: ${record.previousWater}`),
            createElement('div', 'user-detail', `补差: ${record.waterAdjustment || 0}`)
        );
        usersGrid.appendChild(card);
    });

    info.append(heading, usersGrid);
    container.replaceChildren(info);
}

/**
 * 解析水表读数输入
 */
function parseWaterReadings(input) {
    return input
        .split(/[\s,\n]+/)
        .map(s => s.trim())
        .filter(s => s.length > 0)
        .map(s => parseFloat(s))
        .filter(n => !isNaN(n));
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

    const selectedRecords = state.allRecords.filter(r => ids.includes(r.id));
    displayMatchPreview(simulateMatch(selectedRecords, readings));
}

/**
 * 本地模拟匹配（用于预览，所有用水量必须非负）
 */
function simulateMatch(records, readings) {
    const n = records.length;
    let bestMatches = [];
    let minTotalUsage = Infinity;
    let hasValidMatch = false;
    const permutations = generatePermutations(readings);

    for (const perm of permutations) {
        let totalUsage = 0;
        const currentMatches = [];
        let isValid = true;

        for (let i = 0; i < n; i++) {
            const usage = perm[i] - records[i].previousWater + (records[i].waterAdjustment || 0);

            if (usage < 0) {
                isValid = false;
                break;
            }

            totalUsage += usage;
            currentMatches.push({
                record: records[i],
                waterReading: perm[i],
                waterUsage: usage
            });
        }

        if (isValid && totalUsage < minTotalUsage) {
            minTotalUsage = totalUsage;
            hasValidMatch = true;
            bestMatches = currentMatches;
        }
    }

    if (!hasValidMatch) {
        return {
            matches: [],
            totalUsage: 0,
            error: '未找到有效的匹配方案（所有方案都会产生负数用水量）'
        };
    }

    return { matches: bestMatches, totalUsage: minTotalUsage };
}

/**
 * 生成排列组合（使用回溯算法）
 */
function generatePermutations(arr) {
    const result = [];
    const n = arr.length;

    if (n === 0) return result;
    if (n === 1) return [[arr[0]]];

    function backtrack(current, start) {
        if (start === n) {
            result.push([...current]);
            return;
        }

        for (let i = start; i < n; i++) {
            [current[start], current[i]] = [current[i], current[start]];
            backtrack(current, start + 1);
            [current[start], current[i]] = [current[i], current[start]];
        }
    }

    const working = [...arr];
    backtrack(working, 0);
    return result;
}

function createErrorPreview(error) {
    const preview = createElement('div', 'match-preview match-error');
    preview.appendChild(createElement('h4', '', '⚠️ 无法匹配'));

    const errorMessage = createElement('div', 'error-message');
    errorMessage.appendChild(createElement('p', '', error || '未找到有效的匹配方案'));

    const reasons = document.createElement('p');
    reasons.style.cssText = 'margin-top: 10px; color: #666;';
    const label = document.createElement('strong');
    label.textContent = '可能原因：';
    reasons.append(
        label,
        document.createElement('br'),
        document.createTextNode('• 输入的水表读数小于上月读数，会产生负数用水量'),
        document.createElement('br'),
        document.createTextNode('• 请检查输入的读数是否正确'),
        document.createElement('br'),
        document.createTextNode('• 确认上月读数和补差值是否准确')
    );
    errorMessage.appendChild(reasons);
    preview.appendChild(errorMessage);
    return preview;
}

function createMatchPreview(result) {
    const preview = createElement('div', 'match-preview');
    preview.appendChild(createElement('h4', '', '🎯 最优匹配方案（总用水量最小，所有用水量非负）'));

    const totalUsage = createElement('div', 'total-usage');
    const totalUsageLabel = document.createElement('strong');
    totalUsageLabel.textContent = '总用水量：';
    totalUsage.append(totalUsageLabel, document.createTextNode(`${result.totalUsage.toFixed(2)} 吨`));

    const matchResults = createElement('div', 'match-results');
    result.matches.forEach(match => {
        const item = createElement('div', 'match-item');
        const details = createElement('div', 'match-details');
        const usageClass = match.waterUsage < 0 ? 'match-usage negative' : 'match-usage';
        details.append(
            createElement('div', '', `上月: ${match.record.previousWater}`),
            createElement('div', 'match-arrow', '→'),
            createElement('div', 'match-current', `本月: ${match.waterReading}`),
            createElement('div', usageClass, `用量: ${match.waterUsage.toFixed(2)} 吨`)
        );
        item.append(createElement('div', 'match-room', match.record.roomNumber), details);
        matchResults.appendChild(item);
    });

    const actions = createElement('div', 'match-actions');
    const confirmButton = createElement('button', 'btn btn-primary', '✅ 确认并应用');
    confirmButton.type = 'button';
    confirmButton.addEventListener('click', () => window.billingApp.executeSmartMatch());
    actions.appendChild(confirmButton);
    preview.append(totalUsage, matchResults, actions);
    return preview;
}

/**
 * 显示匹配预览
 */
function displayMatchPreview(result) {
    const container = document.getElementById('matchPreview');
    const preview = result.error || result.matches.length === 0
        ? createErrorPreview(result.error)
        : createMatchPreview(result);
    container.replaceChildren(preview);
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
