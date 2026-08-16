/**
 * 额外费用管理模块
 */

import { state, resetExtraFeeCounter, resetBatchExtraFeeCounter } from './config.js';

function createExtraFeeInput(id, name, amount, options) {
    const item = document.createElement('div');
    item.className = 'extra-fee-item';
    item.id = options.itemId(id);

    const row = document.createElement('div');
    row.className = 'form-row';

    const nameGroup = document.createElement('div');
    nameGroup.className = 'form-group';
    const nameInput = document.createElement('input');
    nameInput.type = 'text';
    nameInput.placeholder = '费用名称（如：水管维修费）';
    nameInput.className = options.nameClass;
    nameInput.value = name;
    nameGroup.appendChild(nameInput);

    const amountGroup = document.createElement('div');
    amountGroup.className = 'form-group';
    const amountInput = document.createElement('input');
    amountInput.type = 'number';
    amountInput.placeholder = '金额';
    amountInput.step = '0.01';
    amountInput.className = options.amountClass;
    amountInput.value = amount;
    amountGroup.appendChild(amountInput);

    const removeButton = document.createElement('button');
    removeButton.type = 'button';
    removeButton.className = 'btn btn-danger';
    removeButton.style.cssText = 'padding: 5px 10px;';
    removeButton.textContent = '✕';
    removeButton.dataset.feeId = id;
    removeButton.addEventListener('click', () => options.remove(id));

    row.append(nameGroup, amountGroup, removeButton);
    item.appendChild(row);
    return item;
}

/**
 * 添加额外费用输入框
 * @param {string} name - 费用名称
 * @param {number} amount - 金额
 */
export function addExtraFeeInput(name = '', amount = '') {
    const container = document.getElementById('extraFeesContainer');
    const id = state.extraFeeCounter++;
    container.appendChild(createExtraFeeInput(id, name, amount, {
        itemId: feeId => `extraFee${feeId}`,
        nameClass: 'extra-fee-name',
        amountClass: 'extra-fee-amount',
        remove: removeExtraFeeInput
    }));
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
    container.replaceChildren();
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
    container.appendChild(createExtraFeeInput(id, name, amount, {
        itemId: feeId => `batchExtraFee${feeId}`,
        nameClass: 'batch-extra-fee-name',
        amountClass: 'batch-extra-fee-amount',
        remove: removeBatchExtraFeeInput
    }));
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
    container.replaceChildren();
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
