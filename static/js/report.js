/**
 * 报表生成模块
 */

import { state } from './config.js';
import { fetchGeneratedReport, fetchGeneratedCard } from './api.js';
import { showNotification, getSelectedIds } from './utils.js';

function openGeneratedFile(downloadUrl) {
    if (!downloadUrl || typeof downloadUrl !== 'string') {
        showNotification('服务器未返回有效下载地址', 'error');
        return;
    }

    const target = new URL(downloadUrl, window.location.origin);
    if (target.origin !== window.location.origin || !target.pathname.startsWith('/api/v1/billing/')) {
        showNotification('服务器返回的下载地址无效', 'error');
        return;
    }

    window.open(target.href, '_blank', 'noopener');
}

/**
 * 生成报表（所有或当前筛选）
 */
export async function generateReport() {
    const month = document.getElementById('monthFilter').value;

    if (!month && state.allRecords.length === 0) {
        showNotification('没有数据可生成报表', 'error');
        return;
    }

    const params = {};

    if (month) {
        params.month = month;
    } else {
        const firstMonth = state.allRecords[0]?.billingMonth;
        if (firstMonth) {
            params.month = firstMonth;
        }
    }

    if (state.currentSortOrder) {
        params.sortBy = 'room';
        params.order = state.currentSortOrder;
    }

    const result = await fetchGeneratedReport(params);

    if (result?.success) {
        showNotification('报表生成成功！正在下载...', 'success');
        openGeneratedFile(result.data.downloadUrl);
    } else if (result) {
        showNotification('生成失败: ' + result.error, 'error');
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

    const selectedRecords = state.allRecords.filter(record => ids.includes(record.id));
    const months = [...new Set(selectedRecords.map(record => record.billingMonth).filter(Boolean))];
    const params = { ids };
    if (months.length === 1) {
        params.month = months[0];
    }

    if (state.currentSortOrder) {
        params.sortBy = 'room';
        params.order = state.currentSortOrder;
    }

    const result = await fetchGeneratedReport(params);

    if (result?.success) {
        showNotification(`成功生成${result.data.count}条记录的报表！`, 'success');
        openGeneratedFile(result.data.downloadUrl);
    } else if (result) {
        showNotification('生成失败: ' + result.error, 'error');
    }
}

/**
 * 生成单个卡片
 * @param {number} id - 记录ID
 */
export async function generateSingleCard(id) {
    const result = await fetchGeneratedCard(id);

    if (result?.success) {
        showNotification('卡片生成成功！', 'success');
        openGeneratedFile(result.data.downloadUrl);
    } else if (result) {
        showNotification('生成失败: ' + result.error, 'error');
    }
}
