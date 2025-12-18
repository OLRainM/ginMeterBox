/**
 * 额外费用管理模块
 */

import { state, resetExtraFeeCounter, resetBatchExtraFeeCounter } from './config.js';

/**
 * 添加额外费用输入框
 * @param {string} name - 费用名称
 * @param {number} amount - 金额
 */
export function addExtraFeeInput(name = '', amount = '') {
    const container = document.getElementById('extraFeesContainer');
    const id = state.extraFeeCounter++;
    
    const div = document.createElement('div');
    div.className = 'extra-fee-item';
    div.id = `extraFee${id}`;
    div.innerHTML = `
        <div class="form-row">
            <div class="form-group">
                <input type="text" placeholder="费用名称（如：水管维修费）" 
                       class="extra-fee-name" value="${name}">
            </div>
            <div class="form-group">
                <input type="number" placeholder="金额" step="0.01" 
                       class="extra-fee-amount" value="${amount}">
            </div>
            <button type="button" class="btn btn-danger" onclick="window.billingApp.removeExtraFeeInput(${id})"
                    style="padding: 5px 10px;">
                ✕
            </button>
        </div>
    `;
    
    container.appendChild(div);
}

/**
 * 删除额外费用输入框
 * @param {number} id - 输入框ID
 */
export function removeExtraFeeInput(id) {
    const element = document.getElementById(`extraFee${id}`);
    if (element) {
        element.remove();
    }
}

/**
 * 清空额外费用输入框
 */
export function clearExtraFeeInputs() {
    const container = document.getElementById('extraFeesContainer');
    container.innerHTML = '';
    resetExtraFeeCounter();
}

/**
 * 获取所有额外费用
 * @returns {Array} 额外费用数组
 */
export function getExtraFees() {
    const fees = [];
    const names = document.querySelectorAll('.extra-fee-name');
    const amounts = document.querySelectorAll('.extra-fee-amount');
    
    for (let i = 0; i < names.length; i++) {
        const name = names[i].value.trim();
        const amount = parseFloat(amounts[i].value) || 0;
        
        if (name && amount > 0) {
            fees.push({ name, amount });
        }
    }
    
    return fees;
}

/**
 * 加载额外费用到表单
 * @param {Array} extraFees - 额外费用数组
 */
export function loadExtraFees(extraFees) {
    clearExtraFeeInputs();
    if (extraFees && extraFees.length > 0) {
        extraFees.forEach(fee => {
            addExtraFeeInput(fee.name, fee.amount);
        });
    }
}

// ========== 批量额外费用管理 ==========

/**
 * 添加批量额外费用输入框
 * @param {string} name - 费用名称
 * @param {number} amount - 金额
 */
export function addBatchExtraFeeInput(name = '', amount = '') {
    const container = document.getElementById('batchExtraFeesContainer');
    const id = state.batchExtraFeeCounter++;
    
    const div = document.createElement('div');
    div.className = 'extra-fee-item';
    div.id = `batchExtraFee${id}`;
    div.innerHTML = `
        <div class="form-row">
            <div class="form-group">
                <input type="text" placeholder="费用名称（如：水管维修费）" 
                       class="batch-extra-fee-name" value="${name}">
            </div>
            <div class="form-group">
                <input type="number" placeholder="金额" step="0.01" 
                       class="batch-extra-fee-amount" value="${amount}">
            </div>
            <button type="button" class="btn btn-danger" onclick="window.billingApp.removeBatchExtraFeeInput(${id})"
                    style="padding: 5px 10px;">
                ✕
            </button>
        </div>
    `;
    
    container.appendChild(div);
}

/**
 * 删除批量额外费用输入框
 * @param {number} id - 输入框ID
 */
export function removeBatchExtraFeeInput(id) {
    const element = document.getElementById(`batchExtraFee${id}`);
    if (element) {
        element.remove();
    }
}

/**
 * 清空批量额外费用输入框
 */
export function clearBatchExtraFeeInputs() {
    const container = document.getElementById('batchExtraFeesContainer');
    container.innerHTML = '';
    resetBatchExtraFeeCounter();
}

/**
 * 获取批量额外费用
 * @returns {Array} 额外费用数组
 */
export function getBatchExtraFees() {
    const fees = [];
    const names = document.querySelectorAll('.batch-extra-fee-name');
    const amounts = document.querySelectorAll('.batch-extra-fee-amount');
    
    for (let i = 0; i < names.length; i++) {
        const name = names[i].value.trim();
        const amount = parseFloat(amounts[i].value) || 0;
        
        if (name && amount > 0) {
            fees.push({ name, amount });
        }
    }
    
    return fees;
}
