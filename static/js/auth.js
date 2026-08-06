import { API_BASE_URL } from './config.js';

const loginOverlay = document.getElementById('loginOverlay');
const loginForm = document.getElementById('loginForm');
const passwordInput = document.getElementById('loginPassword');
const loginError = document.getElementById('loginError');

function setError(message = '') {
    loginError.textContent = message;
    loginError.hidden = !message;
}

export function showLogin() {
    setError();
    loginOverlay.hidden = false;
    passwordInput.focus();
}

export function hideLogin() {
    loginOverlay.hidden = true;
    passwordInput.value = '';
    setError();
}

export async function checkSession() {
    try {
        const response = await fetch(`${API_BASE_URL}/auth/session`, {
            credentials: 'same-origin'
        });
        return response.ok;
    } catch (error) {
        return false;
    }
}

export async function logout() {
    await fetch(`${API_BASE_URL}/auth/logout`, {
        method: 'POST',
        credentials: 'same-origin'
    });
    showLogin();
}

export function setupAuth(onAuthenticated) {
    loginForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        setError();

        const submitButton = loginForm.querySelector('button[type="submit"]');
        submitButton.disabled = true;
        try {
            const response = await fetch(`${API_BASE_URL}/auth/login`, {
                method: 'POST',
                credentials: 'same-origin',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ password: passwordInput.value })
            });
            const result = await response.json();
            if (!response.ok || !result.success) {
                setError(result.error || '登录失败，请重试');
                return;
            }
            hideLogin();
            await onAuthenticated();
        } catch (error) {
            setError('无法连接到服务器，请稍后重试');
        } finally {
            submitButton.disabled = false;
        }
    });

    window.addEventListener('auth-required', showLogin);
}
