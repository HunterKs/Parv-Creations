/* Login Page Script */

document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('login-form');
    if (!form) return;

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        await login();
    });
});

/**
 * Handle login form submission
 */
async function login() {
    const form = document.getElementById('login-form');
    const emailInput = document.getElementById('email');
    const passwordInput = document.getElementById('password');
    const rememberCheckbox = document.getElementById('remember');

    const email = emailInput.value.trim();
    const password = passwordInput.value;
    const remember = rememberCheckbox.checked;

    // Reset error states
    emailInput.classList.remove('is-invalid');
    passwordInput.classList.remove('is-invalid');

    // Basic validation
    if (!email) {
        emailInput.classList.add('is-invalid');
        showNotification('Please enter your email', 'error');
        return;
    }

    if (!password) {
        passwordInput.classList.add('is-invalid');
        showNotification('Please enter your password', 'error');
        return;
    }

    // Show loading state on button
    const submitButton = form.querySelector('button[type="submit"]');
    const originalButtonText = submitButton.innerHTML;
    submitButton.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Logging in...';
    submitButton.disabled = true;

    try {
        // Prepare login data
        const loginData = {
            email: email,
            password: password,
            remember: remember
        };

        // Send login request
        const response = await api.post('/auth/login', loginData);

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Invalid email or password');
        }

        // Login successful
        showNotification('Login successful', 'success');

        // Redirect to dashboard
        window.location.href = '/admin/';
    } catch (error) {
        console.error('Login failed:', error);
        showNotification(`Login failed: ${error.message}`, 'error');

        // Reset button state
        submitButton.innerHTML = originalButtonText;
        submitButton.disabled = false;
    } finally {
        // Ensure button is reset in case of success too (though we redirect)
        // Actually, on success we redirect, so this won't be reached
        // But we'll keep it for safety
        submitButton.innerHTML = originalButtonText;
        submitButton.disabled = false;
    }
}