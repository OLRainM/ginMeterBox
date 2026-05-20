/**
 * 配置文件
 * 包含全局配置和状态变量
 */

// API配置
export const API_BASE_URL = 'http://localhost:8080/api/v1';

// 全局状态
export const state = {
    allRecords: [],
    editingId: null,
    currentSortOrder: null, // 'asc' 或 'desc'
    currentMonth: '', // 当前选中的月份
    extraFeeCounter: 0, // 额外费用输入框计数器
    batchExtraFeeCounter: 0 // 批量额外费用输入框计数器
};

// 重置额外费用计数器
export function resetExtraFeeCounter() {
    state.extraFeeCounter = 0;
}

// 重置批量额外费用计数器
export function resetBatchExtraFeeCounter() {
    state.batchExtraFeeCounter = 0;
}
