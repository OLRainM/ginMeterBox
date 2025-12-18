/**
 * 报表生成模块
 */

import { API_BASE_URL, state } from './config.js';
import { showNotification, getSelectedIds } from './utils.js';

/**
 * 生成报表（所有或当前筛选）
 */
export async function generateReport() {
    const month = document.getElementById('monthFilter').value;
    
    if (!month && state.allRecords.length === 0) {
        showNotification('没有数据可生成报表', 'error');
        return;
    }
    
    try {
        let url = `${API_BASE_URL}/billing/report/generate`;
        const params = [];
        
        if (month) {
            params.push(`month=${month}`);
        } else {
            // 使用所有记录的月份（取第一条记录的月份）
            const firstMonth = state.allRecords[0]?.billingMonth;
            if (firstMonth) {
                params.push(`month=${firstMonth}`);
            }
        }
        
        // 添加排序参数
        if (state.currentSortOrder) {
            params.push(`sortBy=room&order=${state.currentSortOrder}`);
        }
        
        if (params.length > 0) {
            url += '?' + params.join('&');
        }
        
        const response = await fetch(url);
        const result = await response.json();
        
        if (result.success) {
            showNotification('报表生成成功！正在下载...', 'success');
            // 下载图片
            window.open(`http://localhost:8080/${result.data.filename}`, '_blank');
        } else {
            showNotification('生成失败: ' + result.error, 'error');
        }
    } catch (error) {
        console.error('生成报表失败:', error);
        showNotification('生成报表失败', 'error');
    }
}

/**
 * 生成选中记录的报表
 */
export async function generateSelectedReport() {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要生成报表的记录', 'error');
        return;
    }
    
    try {
        let url = `${API_BASE_URL}/billing/report/generate?ids=${ids.join(',')}`;
        
        // 添加排序参数
        if (state.currentSortOrder) {
            url += `&sortBy=room&order=${state.currentSortOrder}`;
        }
        
        const response = await fetch(url);
        const result = await response.json();
        
        if (result.success) {
            showNotification(`成功生成${result.data.count}条记录的报表！`, 'success');
            // 下载图片
            window.open(`http://localhost:8080/${result.data.filename}`, '_blank');
        } else {
            showNotification('生成失败: ' + result.error, 'error');
        }
    } catch (error) {
        console.error('生成报表失败:', error);
        showNotification('生成报表失败', 'error');
    }
}

/**
 * 生成单个卡片
 * @param {number} id - 记录ID
 */
export async function generateSingleCard(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/card/${id}`);
        const result = await response.json();
        
        if (result.success) {
            showNotification('卡片生成成功！', 'success');
            // 下载图片
            window.open(`http://localhost:8080/${result.data.filename}`, '_blank');
        } else {
            showNotification('生成失败: ' + result.error, 'error');
        }
    } catch (error) {
        console.error('生成卡片失败:', error);
        showNotification('生成卡片失败', 'error');
    }
}
