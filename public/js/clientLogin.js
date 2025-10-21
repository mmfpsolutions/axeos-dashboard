/**
 * @file This script handles the client-side logic for the login page.
 * It captures form submissions, securely hashes the user's password using the
 * Web Crypto API (SHA-256), and sends the credentials to the server's login endpoint.
 * It also handles displaying success or error feedback to the user.
 */

document.addEventListener('DOMContentLoaded', () => {
    /**
     * Asynchronously hashes a password string using the SHA-256 algorithm.
     * This is a security measure to avoid sending plain-text passwords over the network.
     * It should always be used in conjunction with HTTPS for proper security.
     * @param {string} password The plain-text password to hash.
     * @returns {Promise<string>} A promise that resolves to the SHA-256 hash as a hexadecimal string.
     */
    async function hashPasswordSHA256(password) {
        // 1. Encode the password string into a Uint8Array.
        const encoder = new TextEncoder();
        const data = encoder.encode(password);

        // 2. Hash the data using the Web Crypto API's SHA-256 implementation.
        const hashBuffer = await crypto.subtle.digest('SHA-256', data);

        // 3. Convert the resulting ArrayBuffer hash into a hexadecimal string for transmission.
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const hashHex = hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('');

        return hashHex;
    }

    const loginForm = document.getElementById('loginForm');
    const loginError = document.getElementById('login-error');

    // Ensure the login form exists on the page before adding an event listener.
    if (loginForm) {
        // Attach an event listener to handle the form's submit event.
        loginForm.addEventListener('submit', async (event) => {
            // Prevent the browser's default form submission behavior, which would cause a page reload.
            event.preventDefault(); 

            // Hide any previous error messages.
            loginError.style.display = 'none';

            const username = loginForm.username.value;
            const password = loginForm.password.value;

            try {
                // Hash the password before sending it to the server.
                const hashedPassword = await hashPasswordSHA256(password);

                // Send the login request to the API endpoint.
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ username: username, hashedPassword: hashedPassword }),
                });

                if (response.ok) {
                    // On successful login (HTTP 200-299), the server sets a session cookie.
                    // Redirect the user to the main dashboard page.
                    window.location.href = '/';
                } else {
                    // If the server returns an error (e.g., 401 Unauthorized), display the error message.
                    const result = await response.json();
                    loginError.textContent = result.message || 'An unknown error occurred.';
                    loginError.style.display = 'block';
                }
             
            } catch (error) {
                // Handle network errors or other issues with the fetch request itself.
                console.error('Login request failed:', error);
                loginError.textContent = 'Failed to connect to the server. Please try again later.';
                loginError.style.display = 'block';
            }
        });
    }
    
});