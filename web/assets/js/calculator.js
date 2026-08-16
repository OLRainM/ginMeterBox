/**
 * 计算器模块
 */

/**
 * 显示计算器
 */
export function showCalculator() {
    document.getElementById('calculatorModal').style.display = 'block';
}

/**
 * 关闭计算器
 */
export function closeCalculator() {
    document.getElementById('calculatorModal').style.display = 'none';
}

/**
 * 设置计算器
 */
export function setupCalculator() {
    const calcWaterUsage = document.getElementById('calcWaterUsage');
    const calcWaterPrice = document.getElementById('calcWaterPrice');
    const calcElectricUsage = document.getElementById('calcElectricUsage');
    const calcElectricPrice = document.getElementById('calcElectricPrice');
    const calcManagementFee = document.getElementById('calcManagementFee');
    
    function calculate() {
        const waterUsage = parseFloat(calcWaterUsage.value) || 0;
        const waterPrice = parseFloat(calcWaterPrice.value) || 0;
        const electricUsage = parseFloat(calcElectricUsage.value) || 0;
        const electricPrice = parseFloat(calcElectricPrice.value) || 0;
        const managementFee = parseFloat(calcManagementFee.value) || 0;
        
        const waterCost = waterUsage * waterPrice;
        const electricCost = electricUsage * electricPrice;
        const totalCost = waterCost + electricCost + managementFee;
        
        document.getElementById('calcWaterResult').textContent = waterCost.toFixed(2);
        document.getElementById('calcElectricResult').textContent = electricCost.toFixed(2);
        document.getElementById('calcTotalWater').textContent = waterCost.toFixed(2);
        document.getElementById('calcTotalElectric').textContent = electricCost.toFixed(2);
        document.getElementById('calcTotalCost').textContent = totalCost.toFixed(2);
    }
    
    calcWaterUsage.addEventListener('input', calculate);
    calcWaterPrice.addEventListener('input', calculate);
    calcElectricUsage.addEventListener('input', calculate);
    calcElectricPrice.addEventListener('input', calculate);
    calcManagementFee.addEventListener('input', calculate);
}
