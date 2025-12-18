/**
 * 工具函数模块
 * 包含通用工具函数
 */

/**
 * 显示通知消息
 * @param {string} message - 消息内容
 * @param {string} type - 消息类型 (success/error/info)
 */
export function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 25px;
        background: ${type === 'success' ? '#28a745' : type === 'error' ? '#dc3545' : '#17a2b8'};
        color: white;
        border-radius: 10px;
        box-shadow: 0 5px 15px rgba(0,0,0,0.3);
        z-index: 10000;
        animation: slideIn 0.3s ease;
    `;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

/**
 * 获取选中的记录ID列表
 * @returns {number[]} 选中的ID数组
 */
export function getSelectedIds() {
    const checkboxes = document.querySelectorAll('.record-checkbox:checked');
    return Array.from(checkboxes).map(cb => parseInt(cb.value));
}

/**
 * 清除所有选择
 */
export function clearSelection() {
    document.getElementById('selectAll').checked = false;
    document.querySelectorAll('.record-checkbox').forEach(cb => cb.checked = false);
}

// 添加CSS动画样式
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(400px);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(400px);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);
