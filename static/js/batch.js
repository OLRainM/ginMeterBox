/**
 * 批量操作模块
 */

import { showNotification, getSelectedIds } from './utils.js';
import { batchDeleteRecords, exportToExcel, batchSetExtraFees, batchSetAdjustment } from './api.js';
import { clearSelectionUI } from './ui.js';
import { addBatchExtraFeeInput, clearBatchExtraFeeInputs, getBatchExtraFees } from './extraFee.js';

/**
 * 批量删除记录
 * @param {Function} onSuccess - 成功回调
 */
export async function handleBatchDelete(onSuccess) {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要删除的记录', 'error');
        return;
    }
    
    const confirmMsg = `确定要删除选中的 ${ids.length} 条记录吗？\n\n⚠️ 此操作不可撤销！`;
    if (!confirm(confirmMsg)) {
        return;
    }
    
    const result = await batchDeleteRecords(ids);
    if (result) {
        clearSelectionUI();
        if (onSuccess) onSuccess();
    }
}

/**
 * 导出选中记录为Excel
 */
export async function handleExportToExcel() {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要导出的记录', 'error');
        return;
    }
    
    const result = await exportToExcel(ids);
    
    if (result) {
        showNotification(`成功导出 ${result.count} 条记录到Excel`, 'success');
        
        // 提供下载提示
        const filename = result.file;
        const downloadLink = `${window.location.origin}/${filename}`;
        
        // 创建临时下载链接
        const a = document.createElement('a');
        a.href = downloadLink;
        a.download = filename.split('/').pop();
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        
        // 显示成功消息
        showNotification(`Excel文件已生成，正在下载...`, 'success');
    }
}

/**
 * 显示批量设置额外费用模态框
 */
export function showBatchExtraFeeModal() {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要设置额外费用的记录', 'error');
        return;
    }
    
    // 显示选中的记录数
    document.getElementById('batchSelectedCount').textContent = ids.length;
    
    // 清空之前的输入
    clearBatchExtraFeeInputs();
    
    // 默认添加一个空的输入框
    addBatchExtraFeeInput();
    
    // 显示模态框
    document.getElementById('batchExtraFeeModal').style.display = 'block';
}

/**
 * 关闭批量设置额外费用模态框
 */
export function closeBatchExtraFeeModal() {
    document.getElementById('batchExtraFeeModal').style.display = 'none';
    clearBatchExtraFeeInputs();
}

/**
 * 执行批量设置额外费用
 * @param {Function} onSuccess - 成功回调
 */
export async function executeBatchExtraFee(onSuccess) {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要设置额外费用的记录', 'error');
        return;
    }
    
    const extraFees = getBatchExtraFees();
    
    if (extraFees.length === 0) {
        showNotification('请至少添加一项额外费用', 'error');
        return;
    }
    
    // 获取操作模式
    const mode = document.querySelector('input[name="batchMode"]:checked').value;
    
    const result = await batchSetExtraFees(ids, extraFees, mode);
    
    if (result) {
        closeBatchExtraFeeModal();
        clearSelectionUI();
        if (onSuccess) onSuccess();
    }
}

/**
 * 显示批量设置补差模态框
 */
export function showBatchAdjustmentModal() {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要设置补差的记录', 'error');
        return;
    }
    
    // 显示选中的记录数
    document.getElementById('adjustmentSelectedCount').textContent = ids.length;
    
    // 清空输入框
    document.getElementById('batchWaterAdjustment').value = '';
    document.getElementById('batchElectricAdjustment').value = '';
    
    // 显示模态框
    document.getElementById('batchAdjustmentModal').style.display = 'block';
}

/**
 * 关闭批量设置补差模态框
 */
export function closeBatchAdjustmentModal() {
    document.getElementById('batchAdjustmentModal').style.display = 'none';
}

/**
 * 执行批量设置补差
 * @param {Function} onSuccess - 成功回调
 */
export async function executeBatchAdjustment(onSuccess) {
    const ids = getSelectedIds();
    
    if (ids.length === 0) {
        showNotification('请先选择要设置补差的记录', 'error');
        return;
    }
    
    const waterAdjustmentInput = document.getElementById('batchWaterAdjustment').value;
    const electricAdjustmentInput = document.getElementById('batchElectricAdjustment').value;
    
    // 至少要设置一个值
    if (waterAdjustmentInput === '' && electricAdjustmentInput === '') {
        showNotification('请至少设置一个补差值', 'error');
        return;
    }
    
    const waterAdjustment = waterAdjustmentInput !== '' ? parseFloat(waterAdjustmentInput) : null;
    const electricAdjustment = electricAdjustmentInput !== '' ? parseFloat(electricAdjustmentInput) : null;
    
    const result = await batchSetAdjustment(ids, waterAdjustment, electricAdjustment);
    
    if (result) {
        closeBatchAdjustmentModal();
        clearSelectionUI();
        if (onSuccess) onSuccess();
    }
}
