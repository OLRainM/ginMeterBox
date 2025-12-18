/**
 * 表单处理模块
 * 负责表单的显示、编辑和提交
 */

import { state } from './config.js';
import { fetchRecordById, createRecord, updateRecord } from './api.js';
import { clearExtraFeeInputs, loadExtraFees, getExtraFees } from './extraFee.js';

/**
 * 显示添加表单
 */
export function showAddForm() {
    state.editingId = null;
    document.getElementById('formTitle').textContent = '新增记录';
    document.getElementById('billingForm').reset();
    document.getElementById('recordId').value = '';
    
    // 清空额外费用
    clearExtraFeeInputs();
    
    // 设置默认价格
    const waterPrice = document.getElementById('waterPrice').value;
    const electricPrice = document.getElementById('electricPrice').value;
    
    document.getElementById('formModal').style.display = 'block';
}

/**
 * 编辑记录
 * @param {number} id - 记录ID
 */
export async function editRecord(id) {
    const record = await fetchRecordById(id);
    
    if (record) {
        state.editingId = id;
        
        document.getElementById('formTitle').textContent = '编辑记录';
        document.getElementById('recordId').value = record.id;
        document.getElementById('roomNumber').value = record.roomNumber;
        document.getElementById('billingMonth').value = record.billingMonth;
        document.getElementById('currentWater').value = record.currentWater;
        document.getElementById('previousWater').value = record.previousWater;
        document.getElementById('waterAdjustment').value = record.waterAdjustment;
        document.getElementById('currentElectric').value = record.currentElectric;
        document.getElementById('previousElectric').value = record.previousElectric;
        document.getElementById('electricAdjustment').value = record.electricAdjustment;
        document.getElementById('managementFee').value = record.managementFee;
        
        // 加载额外费用
        loadExtraFees(record.extraFees);
        
        document.getElementById('formModal').style.display = 'block';
    }
}

/**
 * 关闭模态框
 */
export function closeModal() {
    document.getElementById('formModal').style.display = 'none';
    state.editingId = null;
    clearExtraFeeInputs();
}

/**
 * 设置表单处理器
 * @param {Function} onSuccess - 成功后的回调函数
 */
export function setupFormHandler(onSuccess) {
    document.getElementById('billingForm').addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const waterPrice = parseFloat(document.getElementById('waterPrice').value);
        const electricPrice = parseFloat(document.getElementById('electricPrice').value);
        
        const data = {
            roomNumber: document.getElementById('roomNumber').value,
            billingMonth: document.getElementById('billingMonth').value,
            currentWater: parseFloat(document.getElementById('currentWater').value),
            previousWater: parseFloat(document.getElementById('previousWater').value),
            waterAdjustment: parseFloat(document.getElementById('waterAdjustment').value) || 0,
            currentElectric: parseFloat(document.getElementById('currentElectric').value),
            previousElectric: parseFloat(document.getElementById('previousElectric').value),
            electricAdjustment: parseFloat(document.getElementById('electricAdjustment').value) || 0,
            managementFee: parseFloat(document.getElementById('managementFee').value) || 0,
            waterPrice: waterPrice,
            electricPrice: electricPrice,
            extraFees: getExtraFees()
        };
        
        let success = false;
        if (state.editingId) {
            success = await updateRecord(state.editingId, data);
        } else {
            success = await createRecord(data);
        }
        
        if (success) {
            closeModal();
            if (onSuccess) onSuccess();
        }
    });
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const formModal = document.getElementById('formModal');
    const calcModal = document.getElementById('calculatorModal');
    const continueModal = document.getElementById('continueModal');
    const batchExtraFeeModal = document.getElementById('batchExtraFeeModal');
    
    if (event.target === formModal) {
        closeModal();
    }
    if (event.target === calcModal && window.billingApp && window.billingApp.closeCalculator) {
        window.billingApp.closeCalculator();
    }
    if (event.target === continueModal && window.billingApp && window.billingApp.closeContinueModal) {
        window.billingApp.closeContinueModal();
    }
    if (event.target === batchExtraFeeModal && window.billingApp && window.billingApp.closeBatchExtraFeeModal) {
        window.billingApp.closeBatchExtraFeeModal();
    }
}
