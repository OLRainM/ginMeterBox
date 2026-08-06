/**
 * API 调用模块
 * 封装所有与后端的交互
 */

import { API_BASE_URL } from './config.js';
import { showNotification } from './utils.js';

function buildUrl(path, params) {
    const query = new URLSearchParams(params);
    const queryString = query.toString();
    return `${API_BASE_URL}${path}${queryString ? `?${queryString}` : ''}`;
}

async function request(path, { params, ...options } = {}, errorMessage = '请求失败') {
    try {
        const response = await fetch(buildUrl(path, params), {
            credentials: 'same-origin',
            ...options
        });
        let result;

        try {
            result = await response.json();
        } catch (error) {
            throw new Error('服务器返回了无效的 JSON 数据');
        }

        if (response.status === 401) {
            showNotification('未登录或会话已过期', 'error');
            window.dispatchEvent(new CustomEvent('auth-required'));
            return null;
        }
        if (!response.ok) {
            throw new Error(result?.error || `请求失败 (${response.status})`);
        }

        return result;
    } catch (error) {
        console.error(`${errorMessage}:`, error);
        showNotification(errorMessage, 'error');
        return null;
    }
}

function jsonOptions(method, body) {
    return {
        method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(body)
    };
}

/**
 * 加载所有账单记录
 * @param {string|null} sortOrder - 排序方式
 * @returns {Promise<Array>} 账单记录数组
 */
export async function fetchRecords(sortOrder = null) {
    const params = sortOrder ? { sortBy: 'room', order: sortOrder } : undefined;
    const result = await request('/billing', { params }, '加载数据失败');
    return result?.success ? result.data || [] : [];
}

/**
 * 获取单个账单记录
 * @param {number} id - 记录ID
 * @returns {Promise<Object|null>} 账单记录对象
 */
export async function fetchRecordById(id) {
    const result = await request(`/billing/${id}`, {}, '获取记录失败');
    return result?.success ? result.data : null;
}

/**
 * 创建新的账单记录
 * @param {Object} data - 账单数据
 * @returns {Promise<boolean>} 是否成功
 */
export async function createRecord(data) {
    const result = await request('/billing', jsonOptions('POST', data), '保存记录失败');

    if (result?.success) {
        showNotification('添加成功', 'success');
        return true;
    }

    if (result) {
        showNotification('操作失败: ' + result.error, 'error');
    }
    return false;
}

/**
 * 更新账单记录
 * @param {number} id - 记录ID
 * @param {Object} data - 账单数据
 * @returns {Promise<boolean>} 是否成功
 */
export async function updateRecord(id, data) {
    const result = await request(`/billing/${id}`, jsonOptions('PUT', data), '保存记录失败');

    if (result?.success) {
        showNotification('更新成功', 'success');
        return true;
    }

    if (result) {
        showNotification('操作失败: ' + result.error, 'error');
    }
    return false;
}

/**
 * 删除账单记录
 * @param {number} id - 记录ID
 * @returns {Promise<boolean>} 是否成功
 */
export async function deleteRecord(id) {
    const result = await request(`/billing/${id}`, { method: 'DELETE' }, '删除记录失败');

    if (result?.success) {
        showNotification('删除成功', 'success');
        return true;
    }

    if (result) {
        showNotification('删除失败: ' + result.error, 'error');
    }
    return false;
}

/**
 * 批量删除记录
 * @param {number[]} ids - 记录ID数组
 * @returns {Promise<Object|null>} 删除结果
 */
export async function batchDeleteRecords(ids) {
    const result = await request('/billing/batch-delete', jsonOptions('POST', { ids }), '批量删除失败');

    if (result?.success) {
        showNotification(`成功删除 ${result.data.count} 条记录`, 'success');
        return result;
    }

    if (result) {
        showNotification('批量删除失败: ' + result.error, 'error');
    }
    return null;
}

/**
 * 导出选中记录为Excel
 * @param {number[]} ids - 记录ID数组
 * @returns {Promise<Object|null>} 导出结果
 */
export async function exportToExcel(ids) {
    const result = await request('/billing/export-excel', jsonOptions('POST', { ids }), '导出Excel失败');

    if (result?.success) {
        return result;
    }

    if (result) {
        showNotification('导出Excel失败: ' + result.error, 'error');
    }
    return null;
}

/**
 * 批量设置额外费用
 * @param {number[]} ids - 记录ID数组
 * @param {Array} extraFees - 额外费用数组
 * @param {string} mode - 操作模式 (append/replace)
 * @returns {Promise<Object|null>} 操作结果
 */
export async function batchSetExtraFees(ids, extraFees, mode) {
    const result = await request(
        '/billing/batch-extra-fee',
        jsonOptions('POST', { ids, extraFees, mode }),
        '批量设置额外费用失败'
    );

    if (result?.success) {
        const modeText = mode === 'append' ? '追加' : '替换';
        showNotification(`成功为${result.data.count}条记录${modeText}额外费用！`, 'success');
        return result;
    }

    if (result) {
        showNotification('批量设置失败: ' + result.error, 'error');
    }
    return null;
}

/**
 * 获取最新记录（用于自动延续）
 * @param {string} roomNumber - 房号
 * @returns {Promise<Object|null>} 记录对象
 */
export async function fetchLatestRecord(roomNumber) {
    const result = await request(
        `/billing/latest/${encodeURIComponent(roomNumber)}`,
        {},
        '获取数据失败'
    );

    if (result?.success) {
        return result.data;
    }

    if (result) {
        showNotification('未找到该住户的历史记录', 'error');
    }
    return null;
}

/**
 * 单户自动延续
 * @param {string} roomNumber - 房号
 * @param {string} newMonth - 新月份
 * @returns {Promise<boolean>} 是否成功
 */
export async function continueRecord(roomNumber, newMonth) {
    const result = await request('/billing/continue', jsonOptions('POST', { roomNumber, newMonth }), '自动延续失败');

    if (result?.success) {
        showNotification(result.message || result.data?.message || '自动延续成功！', 'success');
        return true;
    }

    if (result) {
        showNotification('自动延续失败: ' + result.error, 'error');
    }
    return false;
}

/**
 * 批量自动延续
 * @param {string[]} roomNumbers - 房号数组
 * @param {string} newMonth - 新月份
 * @returns {Promise<Object|null>} 操作结果
 */
export async function batchContinueRecords(roomNumbers, newMonth) {
    const result = await request(
        '/billing/batch-continue',
        jsonOptions('POST', { roomNumbers, newMonth }),
        '批量自动延续失败'
    );

    if (result?.success) {
        showNotification(`成功为 ${result.data.count} 个住户创建新记录！`, 'success');
        return result;
    }

    if (result) {
        showNotification('批量自动延续失败: ' + result.error, 'error');
    }
    return null;
}

/**
 * 批量设置水电补差
 * @param {number[]} ids - 记录ID数组
 * @param {number|null} waterAdjustment - 水表补差值
 * @param {number|null} electricAdjustment - 电表补差值
 * @returns {Promise<Object|null>} 操作结果
 */
export async function batchSetAdjustment(ids, waterAdjustment, electricAdjustment) {
    const body = { ids };

    if (waterAdjustment !== null && waterAdjustment !== undefined) {
        body.waterAdjustment = waterAdjustment;
    }
    if (electricAdjustment !== null && electricAdjustment !== undefined) {
        body.electricAdjustment = electricAdjustment;
    }

    const result = await request('/billing/batch-adjustment', jsonOptions('POST', body), '批量设置补差失败');

    if (result?.success) {
        showNotification(`✅ ${result.message || result.data?.message}`, 'success');
        return result;
    }

    if (result) {
        showNotification('批量设置补差失败: ' + result.error, 'error');
    }
    return null;
}

/**
 * 智能水表匹配
 * @param {number[]} ids - 记录ID数组
 * @param {number[]} waterReadings - 水表读数数组
 * @returns {Promise<Object|null>} 匹配结果
 */
export async function smartWaterMatch(ids, waterReadings) {
    const result = await request(
        '/billing/smart-water-match',
        jsonOptions('POST', { ids, waterReadings }),
        '智能匹配失败'
    );

    if (result?.success) {
        showNotification(`✅ ${result.message || result.data?.message}`, 'success');

        if (result.data?.matches && result.data.matches.length > 0) {
            console.log('匹配结果:', result.data.matches);
        }

        return result;
    }

    if (result) {
        showNotification('智能匹配失败: ' + result.error, 'error');
    }
    return null;
}

/**
 * 获取指定月份的总表记录
 * @param {string} month - 月份
 * @returns {Promise<Object|null>} 总表记录
 */
export async function fetchTotalMeter(month) {
    const result = await request('/total-meter/month', { params: { month } }, '获取总表记录失败');
    return result?.success ? result.data : null;
}

/**
 * 生成账单报表
 * @param {Object} params - 报表查询参数
 * @returns {Promise<Object|null>} 报表结果
 */
export async function fetchGeneratedReport(params) {
    return request('/billing/report/generate', { params }, '生成报表失败');
}

/**
 * 生成单个账单卡片
 * @param {number} id - 记录ID
 * @returns {Promise<Object|null>} 卡片结果
 */
export async function fetchGeneratedCard(id) {
    return request(`/billing/card/${id}`, {}, '生成卡片失败');
}
