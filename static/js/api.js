/**
 * API 调用模块
 * 封装所有与后端的交互
 */

import { API_BASE_URL } from './config.js';
import { showNotification } from './utils.js';

/**
 * 加载所有账单记录
 * @param {string|null} sortOrder - 排序方式
 * @returns {Promise<Array>} 账单记录数组
 */
export async function fetchRecords(sortOrder = null) {
    try {
        let url = `${API_BASE_URL}/billing`;
        if (sortOrder) {
            url += `?sortBy=room&order=${sortOrder}`;
        }

        const response = await fetch(url);
        const result = await response.json();

        if (result.success) {
            return result.data || [];
        }
        return [];
    } catch (error) {
        console.error('加载数据失败:', error);
        showNotification('加载数据失败', 'error');
        return [];
    }
}

/**
 * 获取单个账单记录
 * @param {number} id - 记录ID
 * @returns {Promise<Object|null>} 账单记录对象
 */
export async function fetchRecordById(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/${id}`);
        const result = await response.json();

        if (result.success) {
            return result.data;
        }
        return null;
    } catch (error) {
        console.error('获取记录失败:', error);
        showNotification('获取记录失败', 'error');
        return null;
    }
}

/**
 * 创建新的账单记录
 * @param {Object} data - 账单数据
 * @returns {Promise<boolean>} 是否成功
 */
export async function createRecord(data) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (result.success) {
            showNotification('添加成功', 'success');
            return true;
        } else {
            showNotification('操作失败: ' + result.error, 'error');
            return false;
        }
    } catch (error) {
        console.error('保存记录失败:', error);
        showNotification('保存记录失败', 'error');
        return false;
    }
}

/**
 * 更新账单记录
 * @param {number} id - 记录ID
 * @param {Object} data - 账单数据
 * @returns {Promise<boolean>} 是否成功
 */
export async function updateRecord(id, data) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });

        const result = await response.json();

        if (result.success) {
            showNotification('更新成功', 'success');
            return true;
        } else {
            showNotification('操作失败: ' + result.error, 'error');
            return false;
        }
    } catch (error) {
        console.error('保存记录失败:', error);
        showNotification('保存记录失败', 'error');
        return false;
    }
}

/**
 * 删除账单记录
 * @param {number} id - 记录ID
 * @returns {Promise<boolean>} 是否成功
 */
export async function deleteRecord(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/${id}`, {
            method: 'DELETE'
        });
        const result = await response.json();

        if (result.success) {
            showNotification('删除成功', 'success');
            return true;
        } else {
            showNotification('删除失败: ' + result.error, 'error');
            return false;
        }
    } catch (error) {
        console.error('删除记录失败:', error);
        showNotification('删除记录失败', 'error');
        return false;
    }
}

/**
 * 批量删除记录
 * @param {number[]} ids - 记录ID数组
 * @returns {Promise<Object|null>} 删除结果
 */
export async function batchDeleteRecords(ids) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/batch-delete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids })
        });

        const result = await response.json();

        if (result.success) {
            showNotification(`成功删除 ${result.count} 条记录`, 'success');
            return result;
        } else {
            showNotification('批量删除失败: ' + result.error, 'error');
            return null;
        }
    } catch (error) {
        console.error('批量删除失败:', error);
        showNotification('批量删除失败', 'error');
        return null;
    }
}

/**
 * 导出选中记录为Excel
 * @param {number[]} ids - 记录ID数组
 * @returns {Promise<Object|null>} 导出结果
 */
export async function exportToExcel(ids) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/export-excel`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids })
        });

        const result = await response.json();

        if (result.success) {
            return result;
        } else {
            showNotification('导出Excel失败: ' + result.error, 'error');
            return null;
        }
    } catch (error) {
        console.error('导出Excel失败:', error);
        showNotification('导出Excel失败', 'error');
        return null;
    }
}

/**
 * 批量设置额外费用
 * @param {number[]} ids - 记录ID数组
 * @param {Array} extraFees - 额外费用数组
 * @param {string} mode - 操作模式 (append/replace)
 * @returns {Promise<Object|null>} 操作结果
 */
export async function batchSetExtraFees(ids, extraFees, mode) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/batch-extra-fee`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids, extraFees, mode })
        });

        const result = await response.json();

        if (result.success) {
            const modeText = mode === 'append' ? '追加' : '替换';
            showNotification(`成功为${result.count}条记录${modeText}额外费用！`, 'success');
            return result;
        } else {
            showNotification('批量设置失败: ' + result.error, 'error');
            return null;
        }
    } catch (error) {
        console.error('批量设置额外费用失败:', error);
        showNotification('批量设置额外费用失败', 'error');
        return null;
    }
}

/**
 * 获取最新记录（用于自动延续）
 * @param {string} roomNumber - 房号
 * @returns {Promise<Object|null>} 记录对象
 */
export async function fetchLatestRecord(roomNumber) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/latest/${roomNumber}`);
        const result = await response.json();

        if (result.success) {
            return result.data;
        } else {
            showNotification('未找到该住户的历史记录', 'error');
            return null;
        }
    } catch (error) {
        console.error('获取数据失败:', error);
        showNotification('获取数据失败', 'error');
        return null;
    }
}

/**
 * 单户自动延续
 * @param {string} roomNumber - 房号
 * @param {string} newMonth - 新月份
 * @returns {Promise<boolean>} 是否成功
 */
export async function continueRecord(roomNumber, newMonth) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/continue`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ roomNumber, newMonth })
        });

        const result = await response.json();

        if (result.success) {
            showNotification(result.message || '自动延续成功！', 'success');
            return true;
        } else {
            showNotification('自动延续失败: ' + result.error, 'error');
            return false;
        }
    } catch (error) {
        console.error('自动延续失败:', error);
        showNotification('自动延续失败', 'error');
        return false;
    }
}

/**
 * 批量自动延续
 * @param {string[]} roomNumbers - 房号数组
 * @param {string} newMonth - 新月份
 * @returns {Promise<Object|null>} 操作结果
 */
export async function batchContinueRecords(roomNumbers, newMonth) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/batch-continue`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ roomNumbers, newMonth })
        });

        const result = await response.json();

        if (result.success) {
            showNotification(`成功为 ${result.count} 个住户创建新记录！`, 'success');
            return result;
        } else {
            showNotification('批量自动延续失败: ' + result.error, 'error');
            return null;
        }
    } catch (error) {
        console.error('批量自动延续失败:', error);
        showNotification('批量自动延续失败', 'error');
        return null;
    }
}

/**
 * 批量设置水电补差
 * @param {number[]} ids - 记录ID数组
 * @param {number|null} waterAdjustment - 水表补差值
 * @param {number|null} electricAdjustment - 电表补差值
 * @returns {Promise<Object|null>} 操作结果
 */
export async function batchSetAdjustment(ids, waterAdjustment, electricAdjustment) {
    try {
        const body = { ids };

        // 只添加非null的值
        if (waterAdjustment !== null && waterAdjustment !== undefined) {
            body.waterAdjustment = waterAdjustment;
        }
        if (electricAdjustment !== null && electricAdjustment !== undefined) {
            body.electricAdjustment = electricAdjustment;
        }

        const response = await fetch(`${API_BASE_URL}/billing/batch-adjustment`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body)
        });

        const result = await response.json();

        if (result.success) {
            showNotification(`✅ ${result.message}`, 'success');
            return result;
        } else {
            showNotification('批量设置补差失败: ' + result.error, 'error');
            return null;
        }
    } catch (error) {
        console.error('批量设置补差失败:', error);
        showNotification('批量设置补差失败', 'error');
        return null;
    }
}

/**
 * 智能水表匹配
 * @param {number[]} ids - 记录ID数组
 * @param {number[]} waterReadings - 水表读数数组
 * @returns {Promise<Object|null>} 匹配结果
 */
export async function smartWaterMatch(ids, waterReadings) {
    try {
        const response = await fetch(`${API_BASE_URL}/billing/smart-water-match`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ ids, waterReadings })
        });

        const result = await response.json();

        if (result.success) {
            showNotification(`✅ ${result.message}`, 'success');

            // 显示匹配详情
            if (result.matches && result.matches.length > 0) {
                console.log('匹配结果:', result.matches);
            }

            return result;
        } else {
            showNotification('智能匹配失败: ' + result.error, 'error');
            return null;
        }
    } catch (error) {
        console.error('智能匹配失败:', error);
        showNotification('智能匹配失败', 'error');
        return null;
    }
}


/**
 * 获取指定月份的总表记录
 * @param {string} month - 月份
 * @returns {Promise<Object|null>} 总表记录
 */
export async function fetchTotalMeter(month) {
    try {
        const response = await fetch(`${API_BASE_URL}/total-meter/month?month=${month}`);
        const result = await response.json();
        if (result.success) return result.data;
        return null;
    } catch (error) {
        return null;
    }
}