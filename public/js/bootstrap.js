/**
 * @file Bootstrap Configuration Page JavaScript - First-time setup form management.
 * 
 * This client-side module handles the interactive bootstrap configuration form
 * including form validation, device management, password confirmation, JWT key
 * generation, and animated user interactions. It provides a complete setup
 * experience for first-time application configuration.
 * 
 * Features:
 * - Dynamic form field management
 * - Real-time password validation
 * - Device URL entry and validation
 * - JWT key generation with animation
 * - Mining Core configuration toggle
 * - Form submission with validation
 * - Auto-redirect after successful setup
 * 
 * @author Scott Walter
 * @version 2.0.0
 * @since 2.0.0
 */

document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('bootstrapForm');
    const enableAuthCheckbox = document.getElementById('enableAuth');
    const authFields = document.getElementById('authFields');
    const enableMiningCoreCheckbox = document.getElementById('enableMiningCore');
    const miningCoreFields = document.getElementById('miningCoreFields');
    const enableCryptoNodeCheckbox = document.getElementById('enableCryptoNode');
    const cryptoNodeFields = document.getElementById('cryptoNodeFields');
    const generateJWTButton = document.getElementById('generateJWT');
    const addDeviceButton = document.getElementById('addDevice');
    const axeosInstancesContainer = document.getElementById('axeosInstances');
    const addMiningCoreButton = document.getElementById('addMiningCore');
    const miningCoreInstancesContainer = document.getElementById('miningCoreInstances');
    const messageDiv = document.getElementById('bootstrap-message');
    const passwordField = document.getElementById('password');
    const confirmPasswordField = document.getElementById('confirmPassword');
    const passwordMatchDiv = document.getElementById('passwordMatch');

    // Generate initial JWT key
    generateRandomJWTKey();

    // Set up password toggle handlers
    setupPasswordToggles();

    // Set up password validation
    setupPasswordValidation();

    // Toggle authentication fields
    enableAuthCheckbox.addEventListener('change', function() {
        if (this.checked) {
            authFields.style.display = 'block';
            document.getElementById('username').required = true;
            document.getElementById('password').required = true;
            document.getElementById('confirmPassword').required = true;
            document.getElementById('jwtKey').required = true;
        } else {
            authFields.style.display = 'none';
            document.getElementById('username').required = false;
            document.getElementById('password').required = false;
            document.getElementById('confirmPassword').required = false;
            document.getElementById('jwtKey').required = false;
        }
    });

    // Toggle mining core fields
    enableMiningCoreCheckbox.addEventListener('change', function() {
        if (this.checked) {
            miningCoreFields.style.display = 'block';
        } else {
            miningCoreFields.style.display = 'none';
        }
    });

    // Toggle crypto node fields
    enableCryptoNodeCheckbox.addEventListener('change', function() {
        if (this.checked) {
            cryptoNodeFields.style.display = 'block';
            document.getElementById('cryptoNodeType').required = true;
            document.getElementById('cryptoNodeName').required = true;
            document.getElementById('cryptoNodeAlgo').required = true;
            document.getElementById('cryptoNodeId').required = true;
            document.getElementById('cryptoNodeRpcIp').required = true;
            document.getElementById('cryptoNodeRpcPort').required = true;
            document.getElementById('cryptoNodeRpcAuth').required = true;
        } else {
            cryptoNodeFields.style.display = 'none';
            document.getElementById('cryptoNodeType').required = false;
            document.getElementById('cryptoNodeName').required = false;
            document.getElementById('cryptoNodeAlgo').required = false;
            document.getElementById('cryptoNodeId').required = false;
            document.getElementById('cryptoNodeRpcIp').required = false;
            document.getElementById('cryptoNodeRpcPort').required = false;
            document.getElementById('cryptoNodeRpcAuth').required = false;
        }
    });

    // Generate JWT key button with animation
    generateJWTButton.addEventListener('click', function() {
        generateRandomJWTKey();
        
        // Add animation class
        this.classList.add('generating');
        this.textContent = 'Generating...';
        
        setTimeout(() => {
            this.classList.remove('generating');
            this.textContent = 'Generate Key';
        }, 800);
    });

    // Add device button
    addDeviceButton.addEventListener('click', function() {
        addDeviceInstance();
    });

    // Add mining core button
    addMiningCoreButton.addEventListener('click', function() {
        addMiningCoreInstance();
    });

    // Form submission
    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        await submitBootstrapForm();
    });

    // Set up initial remove button handlers
    updateRemoveButtonHandlers();
    updateRemoveMiningCoreButtonHandlers();

    /**
     * Generates a random 32-character alphanumeric string for JWT key
     */
    function generateRandomJWTKey() {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let result = '';
        for (let i = 0; i < 32; i++) {
            result += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        document.getElementById('jwtKey').value = result;
    }

    /**
     * Adds a new device instance to the form
     */
    function addDeviceInstance() {
        const deviceCount = axeosInstancesContainer.children.length;
        const deviceInstance = document.createElement('div');
        deviceInstance.className = 'device-instance';
        
        deviceInstance.innerHTML = `
            <div class="form-group">
                <label>Device Name</label>
                <input type="text" name="deviceName" placeholder="AxeOS${deviceCount + 1}" required>
            </div>
            <div class="form-group">
                <label>Device URL</label>
                <input type="url" name="deviceUrl" placeholder="http://192.168.1.10${deviceCount}" required>
            </div>
            <button type="button" class="remove-device btn-danger">Remove</button>
        `;

        axeosInstancesContainer.appendChild(deviceInstance);
        updateRemoveButtonHandlers();
    }

    /**
     * Updates event handlers for remove device buttons
     */
    function updateRemoveButtonHandlers() {
        const removeButtons = document.querySelectorAll('.remove-device');
        const deviceInstances = document.querySelectorAll('.device-instance');

        removeButtons.forEach((button) => {
            // Show/hide remove buttons based on device count
            if (deviceInstances.length > 1) {
                button.style.display = 'inline-block';
            } else {
                button.style.display = 'none';
            }

            // Add click handler
            button.onclick = function() {
                if (deviceInstances.length > 1) {
                    button.parentElement.remove();
                    updateRemoveButtonHandlers();
                }
            };
        });
    }

    /**
     * Adds a new mining core instance to the form
     */
    function addMiningCoreInstance() {
        const instanceCount = miningCoreInstancesContainer.children.length;
        const miningCoreInstance = document.createElement('div');
        miningCoreInstance.className = 'mining-core-instance';

        miningCoreInstance.innerHTML = `
            <div class="form-group">
                <label>Instance Name</label>
                <input type="text" name="miningCoreName" placeholder="Mining Core ${instanceCount + 1}">
            </div>
            <div class="form-group">
                <label>Mining Core URL</label>
                <input type="url" name="miningCoreUrl" placeholder="http://192.168.1.10${instanceCount}:4000">
            </div>
            <button type="button" class="remove-mining-core btn-danger">Remove</button>
        `;

        miningCoreInstancesContainer.appendChild(miningCoreInstance);
        updateRemoveMiningCoreButtonHandlers();
    }

    /**
     * Updates event handlers for remove mining core buttons
     */
    function updateRemoveMiningCoreButtonHandlers() {
        const removeButtons = document.querySelectorAll('.remove-mining-core');
        const miningCoreInstances = document.querySelectorAll('.mining-core-instance');

        removeButtons.forEach((button) => {
            // Show/hide remove buttons based on instance count
            if (miningCoreInstances.length > 1) {
                button.style.display = 'inline-block';
            } else {
                button.style.display = 'none';
            }

            // Add click handler
            button.onclick = function() {
                if (miningCoreInstances.length > 1) {
                    button.parentElement.remove();
                    updateRemoveMiningCoreButtonHandlers();
                } else {
                    // Clear the last instance instead of removing
                    const nameInput = button.parentElement.querySelector('input[name="miningCoreName"]');
                    const urlInput = button.parentElement.querySelector('input[name="miningCoreUrl"]');
                    if (nameInput) nameInput.value = '';
                    if (urlInput) urlInput.value = '';
                }
            };
        });
    }

    /**
     * Collects form data and formats it for submission
     */
    function collectFormData() {
        const formData = {
            title: document.getElementById('title').value,
            port: document.getElementById('port').value,
            enableAuth: document.getElementById('enableAuth').checked ? 'true' : 'false'
        };

        // Add auth-related fields if authentication is enabled
        if (document.getElementById('enableAuth').checked) {
            formData.username = document.getElementById('username').value;
            formData.password = document.getElementById('password').value;
            formData.jwtKey = document.getElementById('jwtKey').value;
            formData.jwtExpiry = document.getElementById('jwtExpiry').value;
        }

        // Collect AxeOS instances
        const deviceInstances = document.querySelectorAll('.device-instance');
        formData.axeosInstances = [];
        
        deviceInstances.forEach(instance => {
            const name = instance.querySelector('input[name="deviceName"]').value;
            const url = instance.querySelector('input[name="deviceUrl"]').value;
            
            if (name && url) {
                formData.axeosInstances.push({ name, url });
            }
        });

        // Add mining core settings if enabled
        formData.enableMiningCore = document.getElementById('enableMiningCore').checked ? 'true' : 'false';
        if (document.getElementById('enableMiningCore').checked) {
            // Collect Mining Core instances
            const miningCoreInstances = document.querySelectorAll('.mining-core-instance');
            formData.miningCoreInstances = [];

            miningCoreInstances.forEach(instance => {
                const name = instance.querySelector('input[name="miningCoreName"]').value;
                const url = instance.querySelector('input[name="miningCoreUrl"]').value;

                if (name && url) {
                    formData.miningCoreInstances.push({ name, url });
                }
            });
        }

        // Add crypto node settings if enabled
        formData.enableCryptoNode = document.getElementById('enableCryptoNode').checked ? 'true' : 'false';
        if (document.getElementById('enableCryptoNode').checked) {
            formData.cryptoNodeType = document.getElementById('cryptoNodeType').value;
            formData.cryptoNodeName = document.getElementById('cryptoNodeName').value;
            formData.cryptoNodeAlgo = document.getElementById('cryptoNodeAlgo').value;
            formData.cryptoNodeId = document.getElementById('cryptoNodeId').value;
            formData.cryptoNodeRpcIp = document.getElementById('cryptoNodeRpcIp').value;
            formData.cryptoNodeRpcPort = document.getElementById('cryptoNodeRpcPort').value;
            formData.cryptoNodeRpcAuth = document.getElementById('cryptoNodeRpcAuth').value;
        }

        return formData;
    }

    /**
     * Submits the bootstrap form
     */
    async function submitBootstrapForm() {
        try {
            showMessage('Validating configuration and testing device connections...', 'info');
            
            const formData = collectFormData();
            
            // Validate required fields only if authentication is enabled
            if (formData.enableAuth === 'true') {
                if (!formData.password) {
                    showMessage('Password is required when authentication is enabled.', 'error');
                    return;
                }
                
                if (!validatePasswords()) {
                    showMessage('Please ensure passwords match and meet requirements.', 'error');
                    return;
                }
            }

            if (formData.axeosInstances.length === 0) {
                showMessage('At least one AxeOS device is required.', 'error');
                return;
            }

            const response = await fetch('/bootstrap', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(formData)
            });

            const result = await response.json();

            if (result.success) {
                showMessage('Configuration created successfully! Redirecting to dashboard...', 'success');
                form.style.display = 'none';
                
                // Wait 2 seconds then redirect to allow the server to switch modes
                setTimeout(() => {
                    window.location.href = '/';
                }, 2000);
            } else {
                showMessage('Validation failed: ' + result.message, 'error');
                // Keep the form visible so user can fix the issues
            }
        } catch (error) {
            console.error('Bootstrap submission error:', error);
            showMessage('An error occurred while creating the configuration. Please try again.', 'error');
        }
    }

    /**
     * Shows a message to the user
     * @param {string} message - The message to display
     * @param {string} type - The type of message ('info', 'success', 'error')
     */
    function showMessage(message, type) {
        messageDiv.textContent = message;
        messageDiv.className = `message ${type}`;
        messageDiv.style.display = 'block';

        // Auto-hide info messages after 5 seconds
        if (type === 'info') {
            setTimeout(() => {
                messageDiv.style.display = 'none';
            }, 5000);
        }
    }

    /**
     * Sets up password visibility toggle functionality
     */
    function setupPasswordToggles() {
        const toggleButtons = document.querySelectorAll('.password-toggle');
        
        toggleButtons.forEach(button => {
            button.addEventListener('click', function() {
                const targetId = this.getAttribute('data-target');
                const targetInput = document.getElementById(targetId);
                
                if (targetInput.type === 'password') {
                    targetInput.type = 'text';
                    this.textContent = '🙈';
                } else {
                    targetInput.type = 'password';
                    this.textContent = '👁️';
                }
            });
        });
    }

    /**
     * Sets up password validation
     */
    function setupPasswordValidation() {
        function checkPasswordMatch() {
            const password = passwordField.value;
            const confirmPassword = confirmPasswordField.value;
            
            if (confirmPassword === '') {
                passwordMatchDiv.style.display = 'none';
                return;
            }
            
            passwordMatchDiv.style.display = 'block';
            
            if (password === confirmPassword) {
                passwordMatchDiv.textContent = '✓ Passwords match';
                passwordMatchDiv.className = 'validation-message success';
            } else {
                passwordMatchDiv.textContent = '✗ Passwords do not match';
                passwordMatchDiv.className = 'validation-message error';
            }
        }
        
        passwordField.addEventListener('input', checkPasswordMatch);
        confirmPasswordField.addEventListener('input', checkPasswordMatch);
    }

    /**
     * Validates that passwords match and meet requirements
     * @returns {boolean} True if passwords are valid
     */
    function validatePasswords() {
        if (!document.getElementById('enableAuth').checked) {
            return true; // Skip validation if auth is disabled
        }
        
        const password = passwordField.value;
        const confirmPassword = confirmPasswordField.value;
        
        return password.length >= 1 && password === confirmPassword;
    }
});