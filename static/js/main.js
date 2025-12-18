/**
 * 主入口文件
 * 协调各个模块的初始化和交互
 */

import { state } from './config.js';
import { showNotification } from './utils.js';
import { fetchRecords, deleteRecord } from './api.js';
import { displayRecords, updateStatistics, populateRoomFilter, toggleSelectAll, clearSelectionUI, updateSelectedStatistics } from './ui.js';
import { showAddForm, editRecord, closeModal, setupFormHandler } from './form.js';
import { applyFilters, clearFilter, sortByRoom } from './filter.js';
import { showCalculator, closeCalculator, setupCalculator } from './calculator.js';
import { generateReport, generateSelectedReport, generateSingleCard } from './report.js';
import { showContinueForm, closeContinueModal, toggleContinueMode, selectAllRooms, deselectAllRooms, previewContinue, executeContinue } from './continue.js';
import { addExtraFeeInput, removeExtraFeeInput, addBatchExtraFeeInput, removeBatchExtraFeeInput } from './extraFee.js';
import { handleBatchDelete, handleExportToExcel, showBatchExtraFeeModal, closeBatchExtraFeeModal, executeBatchExtraFee, showBatchAdjustmentModal, closeBatchAdjustmentModal, executeBatchAdjustment } from './batch.js';
import { showSmartMatchModal, closeSmartMatchModal, previewMatch, executeSmartMatch } from './smartMatch.js';

/**
 * 账单管理应用类
 */
class BillingApp {
    constructor() {
        this.init();
    }

    /**
     * 初始化应用
     */
    async init() {
        // 加载数据
        await this.loadRecords();
        
        // 设置表单处理器
        setupFormHandler(() => this.loadRecords(state.currentSortOrder));
        
        // 设置计算器
        setupCalculator();
        
        // 暴露全局方法供HTML调用
        this.exposeGlobalMethods();
    }

    /**
     * 加载所有记录
     */
    async loadRecords(sortOrder = null) {
        const records = await fetchRecords(sortOrder);
        state.allRecords = records;
        populateRoomFilter();
        
        // 检查是否有活动的筛选条件
        const monthFilter = document.getElementById('monthFilter').value;
        const roomFilter = document.getElementById('roomFilter').value;
        
        if (monthFilter || roomFilter) {
            // 如果有筛选条件，重新应用筛选
            applyFilters();
        } else {
            // 没有筛选条件，显示所有记录
            displayRecords(records);
            updateStatistics(records);
        }
    }

    /**
     * 删除记录
     */
    async deleteRecord(id) {
        if (!confirm('确定要删除这条记录吗？')) {
            return;
        }
        
        const success = await deleteRecord(id);
        if (success) {
            await this.loadRecords(state.currentSortOrder);
        }
    }

    /**
     * 按房号排序
     */
    async sortByRoom(order) {
        await sortByRoom(order, (msg) => showNotification(msg, 'success'));
        populateRoomFilter();
    }

    /**
     * 暴露全局方法供HTML内联事件调用
     */
    exposeGlobalMethods() {
        window.billingApp = {
            // 记录管理
            showAddForm,
            editRecord: (id) => editRecord(id),
            deleteRecord: (id) => this.deleteRecord(id),
            closeModal,
            
            // 数据加载
            loadRecords: () => this.loadRecords(),
            
            // 筛选排序
            applyFilters,
            clearFilter,
            filterByMonth: applyFilters, // 向后兼容
            sortByRoom: (order) => this.sortByRoom(order),
            
            // 选择
            toggleSelectAll,
            clearSelection: clearSelectionUI,
            
            // 计算器
            showCalculator,
            closeCalculator,
            
            // 报表
            generateReport,
            generateSelectedReport,
            generateSingleCard,
            
            // 自动延续
            showContinueForm,
            closeContinueModal,
            toggleContinueMode,
            selectAllRooms,
            deselectAllRooms,
            previewContinue,
            executeContinue: () => executeContinue(() => this.loadRecords(state.currentSortOrder)),
            
            // 额外费用
            addExtraFeeInput,
            removeExtraFeeInput,
            addBatchExtraFeeInput,
            removeBatchExtraFeeInput,
            
            // 批量操作
            batchDeleteRecords: () => handleBatchDelete(() => this.loadRecords(state.currentSortOrder)),
            exportToExcel: handleExportToExcel,
            showBatchExtraFeeModal,
            closeBatchExtraFeeModal,
            executeBatchExtraFee: () => executeBatchExtraFee(() => this.loadRecords(state.currentSortOrder)),
            showBatchAdjustmentModal,
            closeBatchAdjustmentModal,
            executeBatchAdjustment: () => executeBatchAdjustment(() => this.loadRecords(state.currentSortOrder)),
            
            // 智能匹配
            showSmartMatchModal,
            closeSmartMatchModal,
            previewMatch,
            executeSmartMatch: () => executeSmartMatch(() => this.loadRecords(state.currentSortOrder))
        };
    }
}

// 页面加载完成后初始化应用
document.addEventListener('DOMContentLoaded', () => {
    new BillingApp();
});
