/**
 * @file AxeOS Dashboard Client-Side Application Logic.
 * 
 * This module provides the complete client-side functionality for the AxeOS
 * Dashboard application. It manages the dynamic dashboard interface, device
 * data polling, user interactions, and real-time updates.
 * 
 * Core Features:
 * - Real-time device data fetching and display
 * - Dynamic navigation menu generation
 * - Interactive device detail panels
 * - Responsive data visualization with progress bars
 * - Modal-based configuration management
 * - Mining pool integration display
 * - Authentication and session management
 * - Auto-refresh with configurable intervals
 * - Error handling and user feedback
 * 
 * Data Flow:
 * 1. Initial load fetches all device and configuration data
 * 2. Navigation menu populated with discovered devices
 * 3. Summary view displays aggregated statistics
 * 4. Device selection shows detailed information
 * 5. Periodic polling keeps data current
 * 6. User actions trigger API calls and UI updates
 * 
 * @author Scott Walter
 * @version 2.0.0
 * @since 1.0.0
 */

document.addEventListener('DOMContentLoaded', () => {
    const timestampSpan = document.getElementById('timestamp');
    const miningCoreDetailsDiv = document.getElementById('mining-core-details');
    const refreshIcon = document.getElementById('refresh-icon');
    const configIcon = document.getElementById('config-icon');


    let minerData = [];
    let displayFieldsConfig = []; // Stores the display_fields from config.json for miners.
    let miningCoreData = null; // Stores the data for the Mining Core instance.
    let miningCoreDisplayFields = []; // Stores the display_fields from config.json for Mining Core.
    let cryptoNodeData = null; // Stores the data for Crypto Nodes.
    let disableSettings=true;
    let disableConfigurations=true;
    let disableAuthentication=false;
    let miningCoreEnabled=false;
    // Define the ASIC Temp, VR Temp and Fan Speed progress bar color limits (green, yellow, red)
    let ASICTempMap = {
        green: 65,
        yellow: 70,
        red: 75,
     }
     let VRTempMap = {
        green: 70,
        yellow: 85,
        red: 100,
     }
     let FanSpeedMap = {
        green: 80,
        yellow: 95,
        red: 100,
     }



    if (refreshIcon) {
        refreshIcon.addEventListener('click', () => {
            location.reload();
        });
    }

    if (configIcon) {
        configIcon.addEventListener('click', () => {
            modalService.openConfigModal();
        });
    }

    // Check for configuration migration on page load
    checkConfigurationMigration();

    // Configuration button will be added after data is loaded and we know the disable_configurations setting

    // --- Retrieve and Parse Data via Fetch ---
    fetch('/api/systems/info')
        .then(response => {
            if (!response.ok) {
                // If the response is not OK (e.g., 404, 500), throw an error.
                const errorMessage = `HTTP error! Status: ${response.status} - ${response.statusText}`;
                console.error(errorMessage);
                miningCoreDetailsDiv.innerHTML = `<p style="color: red;">Error loading device data: ${errorMessage}. Please check the server and refresh.</p>`;
                throw new Error(errorMessage); // Propagate error to the catch block
            }
            return response.json(); // Parse the response body as JSON.
        })
        .then(embedded => {
            minerData = embedded.minerData || [];
            displayFieldsConfig = embedded.displayFields || [];
            miningCoreData = embedded.miningCoreData;
            miningCoreDisplayFields = embedded.miningCoreDisplayFields || [];
            cryptoNodeData = embedded.cryptoNodeData;
            disableSettings = embedded.disable_settings;
            disableConfigurations = embedded.disable_configurations;
            disableAuthentication = embedded.disable_authentication;
            miningCoreEnabled = embedded.mining_core_enabled;

            // Sort data by hostname for a consistent and predictable menu order.
            minerData.sort((a, b) => (a.hostname || a.id).localeCompare(b.hostname || b.id));

            // Update the "Last updated" timestamp on the client side to reflect when the data was fetched.
            if (timestampSpan) {
                timestampSpan.textContent = new Date().toLocaleString();
            }

            // Add logout button based on disable_authentication setting
            addLogoutButton();
            
            // Configuration icon is now part of the header

            // Initialize the dashboard after data is successfully fetched
            displayDashboard();
        })
        .catch(error => {
            console.error('Error fetching or parsing embedded data:', error);
            // Display a user-friendly error message if fetch fails or JSON parsing fails.
            if (!miningCoreDetailsDiv.innerHTML.includes('Error loading device data')) { // Avoid duplicate error messages
                miningCoreDetailsDiv.innerHTML = `<p style="color: red;">Failed to load device data: ${error.message}. Please refresh the page.</p>`;
            }
        });

    // --- Helper Functions ---

    /**
     * Attaches event listeners to Chart buttons in the summary view
     */
    function attachChartButtonEventListeners() {
        const chartButtons = document.querySelectorAll('.chart-button');
        chartButtons.forEach(buttonElement => {
            buttonElement.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const instanceId = buttonElement.getAttribute('data-instance-id');
                if (instanceId && window.statisticsModal) {
                    window.statisticsModal.openStatisticsModal(instanceId);
                }
            });
        });
    }

    /**
     * Attaches event listeners to Restart, Settings, and Info buttons in the summary view
     */
    function attachRestartAndSettingsButtonEventListeners() {
        // Restart button event listeners
        const restartButtons = document.querySelectorAll('.restart-button');
        restartButtons.forEach(buttonElement => {
            buttonElement.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const instanceId = buttonElement.getAttribute('data-instance-id');
                if (instanceId) {
                    modalService.openConfirmModal(
                        'Confirm Restart',
                        `Are you sure you want to restart instance "${instanceId}"?`,
                        async () => {
                            try {
                                const response = await fetch(`/api/instance/service/restart?instanceId=${instanceId}`, {
                                    method: 'POST'
                                });
                                const result = await response.json();
                                if (response.ok) {
                                    alert(`Instance "${instanceId}" is restarting.`);
                                    setTimeout(() => location.reload(), 2000); // Refresh to see updated status
                                } else {
                                    alert(`Error restarting instance: ${result.message || 'Unknown error'}`);
                                }
                            } catch (error) {
                                console.error('Restart request failed:', error);
                                alert('Failed to send restart command. See console for details.');
                            }
                        }
                    );
                }
            });
        });

        // Settings button event listeners
        const settingsButtons = document.querySelectorAll('.settings-button');
        settingsButtons.forEach(buttonElement => {
            buttonElement.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const instanceId = buttonElement.getAttribute('data-instance-id');
                if (instanceId) {
                    // Find the miner data for this instance
                    const selectedData = minerData.find(m => m.id === instanceId);
                    if (selectedData) {
                        modalService.openSettingsModal(selectedData);
                    }
                }
            });
        });

        // Info button event listeners
        const infoButtons = document.querySelectorAll('.info-button');
        infoButtons.forEach(buttonElement => {
            buttonElement.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                const instanceId = buttonElement.getAttribute('data-instance-id');
                if (instanceId) {
                    // Find the miner data for this instance
                    const selectedData = minerData.find(m => m.id === instanceId);
                    if (selectedData) {
                        openMinerInfoModal(selectedData);
                    }
                }
            });
        });
    }

    /**
     * Opens a modal displaying detailed information for a specific miner
     * @param {object} minerData - The miner data object
     */
    function openMinerInfoModal(minerData) {
        const existingModal = document.getElementById('miner-info-modal');
        if (existingModal) existingModal.remove();

        // Generate detailed HTML using the same formatting as device details
        const detailsHtml = generateDeviceDetailsHtml(minerData);

        const modalHtml = `
            <div id="miner-info-modal" class="modal">
                <div class="modal-content" style="max-width: 800px; max-height: 80vh; overflow-y: auto;">
                    <span class="close-button">&times;</span>
                    <div class="miner-info-content">
                        ${detailsHtml}
                    </div>
                </div>
            </div>`;

        document.body.insertAdjacentHTML('beforeend', modalHtml);

        const modal = document.getElementById('miner-info-modal');
        const closeModal = () => modal.remove();

        modal.querySelector('.close-button').addEventListener('click', closeModal);
        window.addEventListener('click', (event) => {
            if (event.target === modal) closeModal();
        });
    }

    /**
     * Saves the collapsed state of a section to localStorage
     */
    function saveSectionState(sectionId, isCollapsed) {
        try {
            const savedStates = JSON.parse(localStorage.getItem('axeos-section-states') || '{}');
            savedStates[sectionId] = isCollapsed;
            localStorage.setItem('axeos-section-states', JSON.stringify(savedStates));
        } catch (error) {
            console.warn('Failed to save section state:', error);
        }
    }

    /**
     * Gets the saved collapsed state of a section from localStorage
     */
    function getSectionState(sectionId) {
        try {
            const savedStates = JSON.parse(localStorage.getItem('axeos-section-states') || '{}');
            return savedStates[sectionId] || false; // Default to expanded (false)
        } catch (error) {
            console.warn('Failed to get section state:', error);
            return false; // Default to expanded
        }
    }

    /**
     * Restores saved collapsed states for all sections
     */
    function restoreSavedSectionStates() {
        const collapseButtons = document.querySelectorAll('.collapse-button');
        collapseButtons.forEach(button => {
            const targetId = button.getAttribute('data-target');
            const targetElement = document.getElementById(targetId);

            if (targetElement) {
                const isCollapsed = getSectionState(targetId);

                if (isCollapsed) {
                    targetElement.classList.add('collapsed');
                    button.textContent = '+';
                } else {
                    targetElement.classList.remove('collapsed');
                    button.textContent = '−';
                }
            }
        });
    }

    /**
     * Attaches event listeners to collapse/expand buttons
     */
    function attachCollapseButtonEventListeners() {
        const collapseButtons = document.querySelectorAll('.collapse-button');
        collapseButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();

                const targetId = button.getAttribute('data-target');
                const targetElement = document.getElementById(targetId);

                if (targetElement) {
                    const isCollapsed = targetElement.classList.contains('collapsed');

                    if (isCollapsed) {
                        // Expand
                        targetElement.classList.remove('collapsed');
                        button.textContent = '−';
                        saveSectionState(targetId, false);
                    } else {
                        // Collapse
                        targetElement.classList.add('collapsed');
                        button.textContent = '+';
                        saveSectionState(targetId, true);
                    }
                }
            });
        });
    }

    /**
     * Safely retrieves a value from the nested structure of the mining core pool data.
     * It checks networkStats, poolStats, and the top-level of the pool object.
     * @param {string} fieldKey The key of the value to retrieve.
     * @param {object} pool The pool data object.
     * @returns {*} The found value, or undefined if not found.
     */
    function getNestedMiningCoreValue(fieldKey, pool) {
        if (pool.networkStats && pool.networkStats.hasOwnProperty(fieldKey)) {
            return pool.networkStats[fieldKey];
        }
        if (pool.poolStats && pool.poolStats.hasOwnProperty(fieldKey)) {
            return pool.poolStats[fieldKey];
        }
        if (pool.hasOwnProperty(fieldKey)) {
            return pool[fieldKey];
        }
        return undefined; // Return undefined to be handled by the formatting function.
    }


    /**
     * Safely formats a number to a specified number of decimal places, or returns 'N/A'.
     * @param {number|string} value The value to format.
     * @param {number} [digits=2] The number of decimal places.
     * @returns {string}
     */
    function safeToFixed(value, digits = 2) {
        return typeof value === 'number' && !isNaN(value) ? value.toFixed(digits) : 'N/A';
    }

    /**
     * Formats uptime from seconds to human-readable string.
     * @param {number} seconds The total uptime in seconds.
     * @returns {string}
     */
    function formatUptime(seconds) {
        if (typeof seconds !== 'number' || isNaN(seconds) || seconds < 0) {
            return 'N/A';
        }
        const days = Math.floor(seconds / (3600 * 24));
        seconds %= (3600 * 24);
        const hours = Math.floor(seconds / 3600);
        seconds %= 3600;
        const minutes = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);

        let parts = [];
        if (days > 0) parts.push(`${days}d`);
        if (hours > 0) parts.push(`${hours}h`);
        if (minutes > 0) parts.push(`${minutes}m`);
        if (secs > 0 || parts.length === 0) parts.push(`${secs}s`); // Ensure at least seconds are shown, even for 0s uptime.

        return parts.join(' ');
    }

    /**
     * Converts a raw hashrate (in H/s) into a human-readable format with appropriate units.
     * @param {number} hashrate - The hashrate in H/s.
     * @returns {string} The formatted hashrate string (e.g., "10.5 TH/s").
     */
    function formatHashrate(hashrate) {
        if (typeof hashrate !== 'number' || isNaN(hashrate) || hashrate < 0) {
            return 'N/A';
        }

        const units = ['H/s', 'KH/s', 'MH/s', 'GH/s', 'TH/s', 'PH/s', 'EH/s', 'ZH/s'];
        let i = 0;
        while (hashrate >= 1000 && i < units.length - 1) {
            hashrate /= 1000;
            i++;
        }
        return `${hashrate.toFixed(2)} ${units[i]}`;
    }

    /**
     * Converts a device hashrate (assumed to be in GH/s) into a human-readable format.
     * @param {number} hashrate - The hashrate in GH/s.
     * @returns {string} The formatted hashrate string (e.g., "500.00 MH/s", "1.63 TH/s").
     */
    function formatDeviceHashrate(hashrate) {
        if (typeof hashrate !== 'number' || isNaN(hashrate) || hashrate < 0) {
            return 'N/A';
        }

        if (hashrate < 1) { // Less than 1 GH/s, show as MH/s
            return `${(hashrate * 1000).toFixed(2)} Mh/s`;
        }

        const units = ['Gh/s', 'Th/s', 'Ph/s', 'Eh/s', 'Zh/s'];
        let i = 0;
        while (hashrate >= 1000 && i < units.length - 1) {
            hashrate /= 1000;
            i++;
        }
        return `${hashrate.toFixed(2)} ${units[i]}`;
    }

    /**
     * Converts a large number (like difficulty) into a human-readable format with metric prefixes.
     * @param {number} value - The number to format.
     * @returns {string} The formatted number string (e.g., "10.5 T").
     */
    function formatLargeNumber(value) {
        if (typeof value !== 'number' || isNaN(value) || value < 0) {
            return 'N/A';
        }

        const units = ['', 'K', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y'];
        let i = 0;
        while (value >= 1000 && i < units.length - 1) {
            value /= 1000;
            i++;
        }
        return `${value.toFixed(2)} ${units[i]}`.trim();
    }

    /**
     * Generates HTML for a horizontal progress bar.
     * @param {number} value - The current value.
     * @param {number} maxValue - The maximum value for the scale.
     * @param {string} fillColor - The color of the filled part of the bar.
     * @returns {string} The HTML string for the progress bar.
     */
    function generateProgressBarHtml(value, maxValue, fillColor) {
        if (typeof value !== 'number' || isNaN(value)) {
            return 'N/A';
        }
        const percentage = Math.min(100, Math.max(0, (value / maxValue) * 100));
        return `
            <div style="width: 100px; height: 10px; background-color: #e0e0e0; border: 1px solid #ccc; border-radius: 3px; overflow: hidden; display: inline-block; vertical-align: middle;">
                <div style="width: ${percentage}%; height: 100%; background-color: ${fillColor};"></div>
            </div>
            <span style="margin-left: 5px;">${value.toFixed(1)} / ${maxValue}</span>
        `;
    }


    /**
     * Applies specific formatting based on the field key.
     * @param {string} fieldKey - The key of the field (e.g., 'hashRate', 'uptimeSeconds').
     * @param {*} value - The raw value of the field.
     * @param {object} [data] - Optional full data object for fields that need access to other fields.
     * @returns {string} The formatted value.
     */
    function formatFieldValue(fieldKey, value, data) {
        if (value === undefined || value === null || value === '') {
            return 'N/A';
        }

        switch (fieldKey) {
            case 'uptimeSeconds':
                return formatUptime(value);
            case 'hashRate':
            case 'expectedHashrate':
                return formatDeviceHashrate(value);
            case 'current':
            case 'temptarget':
            case 'poolDifficulty':
                return typeof value === 'number' && !isNaN(value) ? Math.round(value).toString() : 'N/A';
            case 'power':
                if (typeof value !== 'number' || isNaN(value)) {
                    return 'N/A';
                }
                // Max power is assumed to be 40W for progress bar scaling.
                const powerPercentage = Math.min(100, Math.max(0, (value / 40) * 100));
                return `
                    <div style="width: 100px; height: 10px; background-color: #e0e0e0; border: 1px solid #ccc; border-radius: 3px; overflow: hidden; display: inline-block; vertical-align: middle;">
                        <div style="width: ${powerPercentage}%; height: 100%; background-color: green;"></div>
                    </div>
                    <span style="margin-left: 5px;">${value.toFixed(3)} / 40</span>
                `;
            case 'voltage':
                const voltageInVolts = value / 1000; // Convert millivolts to volts
                if (typeof voltageInVolts !== 'number' || isNaN(voltageInVolts)) {
                    return 'N/A';
                }
                // Max voltage is assumed to be 6V for progress bar scaling.
                const voltagePercentage = Math.min(100, Math.max(0, (voltageInVolts / 6) * 100));
                return `
                    <div style="width: 100px; height: 10px; background-color: #e0e0e0; border: 1px solid #ccc; border-radius: 3px; overflow: hidden; display: inline-block; vertical-align: middle;">
                        <div style="width: ${voltagePercentage}%; height: 100%; background-color: green;"></div>
                    </div>
                    <span style="margin-left: 5px;">${voltageInVolts.toFixed(3)} / 6</span>
                `;
            case 'fanspeed':
                // Color-code the fan speed bar to indicate potential stress.
                let fanSpeedFillColor;
                if (value <= FanSpeedMap.green) {
                    fanSpeedFillColor = 'green';
                } else if (value >= FanSpeedMap.green && value <= FanSpeedMap.yellow) {
                    fanSpeedFillColor = 'yellow';
                } else { // 96 to 100
                    fanSpeedFillColor = 'red';
                }
                if (typeof value !== 'number' || isNaN(value)) {
                    return 'N/A';
                }
                const fanSpeedPercentage = Math.min(100, Math.max(0, (value / 100) * 100));
                return `
                    <div style="width: 100px; height: 10px; background-color: #e0e0e0; border: 1px solid #ccc; border-radius: 3px; overflow: hidden; display: inline-block; vertical-align: middle;">
                        <div style="width: ${fanSpeedPercentage}%; height: 100%; background-color: ${fanSpeedFillColor};"></div>
                    </div>
                    <span style="margin-left: 5px;">${Math.round(value)} / 100</span>
                `;
            case 'coreVoltageActual':
                const coreVoltageActualInVolts = value / 1000; // Convert millivolts to volts
                if (typeof coreVoltageActualInVolts !== 'number' || isNaN(coreVoltageActualInVolts)) {
                    return 'N/A';
                }
                // Max core voltage is assumed to be 1.5V for progress bar scaling.
                const coreVoltageActualPercentage = Math.min(100, Math.max(0, (coreVoltageActualInVolts / 1.5) * 100));
                return `
                    <div style="width: 100px; height: 10px; background-color: #e0e0e0; border: 1px solid #ccc; border-radius: 3px; overflow: hidden; display: inline-block; vertical-align: middle;">
                        <div style="width: ${coreVoltageActualPercentage}%; height: 100%; background-color: green;"></div>
                    </div>
                    <span style="margin-left: 5px;">${coreVoltageActualInVolts.toFixed(3)} / 1.5</span>
                `;
            case 'frequency':
                if (typeof value !== 'number' || isNaN(value)) {
                    return 'N/A';
                }
                // Max frequency is assumed to be 1000MHz for progress bar scaling.
                const frequencyPercentage = Math.min(100, Math.max(0, (value / 1000) * 100));
                return `
                    <div style="width: 100px; height: 10px; background-color: #e0e0e0; border: 1px solid #ccc; border-radius: 3px; overflow: hidden; display: inline-block; vertical-align: middle;">
                        <div style="width: ${frequencyPercentage}%; height: 100%; background-color: green;"></div>
                    </div>
                    <span style="margin-left: 5px;">${Math.round(value)} / 1000</span>
                `;
            case 'temp':
                // Color-code the temperature bar to indicate potential overheating.
                let tempFillColor;
                if (value <= ASICTempMap.green) {
                    tempFillColor = 'green';
                } else if (value > ASICTempMap.green && value <= ASICTempMap.yellow) {
                    tempFillColor = 'yellow';
                } else { // 71 to 75
                    tempFillColor = 'red';
                }
                return generateProgressBarHtml(value, 75, tempFillColor); // Max temp is assumed to be 75°C for scaling.
            case 'vrTemp':
                // Color-code the VRM temperature bar to indicate potential overheating.
                let vrTempFillColor;
                if (value <= VRTempMap.green) {
                    vrTempFillColor = 'green';
                } else if (value >= VRTempMap.green && value <= VRTempMap.yellow) {
                    vrTempFillColor = 'yellow';
                } else { // 86 to 100
                    vrTempFillColor = 'red';
                }
                return generateProgressBarHtml(value, 100, vrTempFillColor); // Max VRM temp is assumed to be 100°C for scaling.
            case 'blockReward': // Keep blockReward here if it's a number needing fixed-point
                return safeToFixed(Number(value));
            case 'networkHashrate':
                return formatHashrate(Number(value));
            case 'poolHashrate':
                return formatHashrate(Number(value));
            case 'networkDifficulty':
                return formatLargeNumber(Number(value));
            case 'totalPaid':
                return safeToFixed(Number(value), 3);
            case 'totalBlocks':
            case 'totalConfirmedBlocks':
            case 'totalPendingBlocks':
            case 'connectedMiners':
            case 'blockHeight':
            case 'connectedPeers':
            case 'nodeVersion':
                // Return as a string to avoid potential precision issues with very large numbers.
                return String(value);
            case 'wifiRSSI':
                return `${value} dBm`;
            case 'overheat_mode':
                const overheatStatus = value ? 'Enabled' : 'Disabled';
                const overheatColor = value ? 'red' : 'green';
                return `<span style="color: ${overheatColor}; font-weight: bold;">${overheatStatus}</span>`;
            case 'isUsingFallbackStratum':
                const fallbackStatus = value ? 'Enabled' : 'Disabled';
                const fallbackColor = value ? 'red' : 'green';
                return `<span style="color: ${fallbackColor}; font-weight: bold;">${fallbackStatus}</span>`;

            case 'lastNetworkBlockTime':
            case 'lastPoolBlockTime':
                // Directly parse ISO 8601 string, no need to multiply by 1000
                const date = new Date(value); // The server provides a standard ISO 8601 string.
                // Check if the date is valid before attempting to format it.
                return isNaN(date.getTime()) ? 'Invalid Date' : date.toLocaleString();
            case 'sharesRejected':
                if (data && typeof data.sharesAccepted === 'number' && data.sharesAccepted > 0) {
                    const percentage = ((value / data.sharesAccepted) * 100).toFixed(2);
                    return `${value} (${percentage}%)`;
                }
                return String(value);
            case 'sharesRejectedReasons':
                if (Array.isArray(value) && value.length > 0) {
                    return value.map(reason => {
                        const message = reason.message === 'unknown' ? 'Duplicate' : reason.message;
                        return `${message}: ${reason.count}`;
                    }).join(', ');
                }
                return 'N/A';
            default:
                return String(value);
        }
    }
    /**
     * Generate the color for a field that has a limits map (green, yellow, red)
     * 
     */
    function getLimitColor(value, limits) {
        if (value <= limits.green) {
            return 'green';
        } else if (value >= limits.green && value <= limits.yellow) {
            return 'yellow';
        } else { 
            return 'red';
        }
    }
    /**
     * Generates the detailed HTML for mining core summary in the right pane.
     * @param {Array} data - The mining core data array (array of instances, each with pools).
     * @param {Array<object>} displayFields - The display_fields configuration for mining core.
     * @returns {string} The HTML string for the mining core details.
     */
    function generateMiningCoreDetailsHtml(data, displayFields) {
        // If we have no miners at all, show an appropriate message
        if (!minerData || minerData.length === 0) {
            return '<p>Loading Mining Dashboard Data. Please wait.</p>';
        }

        // Start with no title, just the content
        let allPoolsHtml = '';
        //Show date / time of last update
         allPoolsHtml += `<div class="mining-pool-summary-card">`;
        allPoolsHtml += '<h3>Status Timestamp</h3>';
        allPoolsHtml += `<div class="details-grid">`;
        allPoolsHtml += `<strong>Last Updated:</strong> <span>${new Date().toLocaleString()}</span>`;
        allPoolsHtml += `</div>`; // Close details-grid for timestamp
        allPoolsHtml += `</div>`; // Close mining-pool-summary-card for timestamp


        // Show each individual miner's status, regardless of whether they are part of a pool.
       // allPoolsHtml += `<div class="mining-pool-summary-card">`; // Container for individual miner status
        allPoolsHtml += `<div class="individual-miner-summary-card">`; // Container for individual miner status
        allPoolsHtml += '<h3><span class="collapse-button" data-target="individual-miner-content">−</span> Individual Miner Status</h3>';
        allPoolsHtml += '<div id="individual-miner-content" class="collapsible-content">';
        allPoolsHtml += '<div class="miner-cards-container">'; // New container for responsive card layout
        // Loop through each miner's data and generate HTML.
            if ( minerData && minerData.length > 0) {
                minerData.forEach(miner => {
                    allPoolsHtml += '<div class="miner-card">'; // Individual card wrapper
                    if (miner.status === 'Error') {
                        // Display the miner's name and its error status.
                        allPoolsHtml += `<h4><span class="status-indicator status-error" style="margin-right: 8px;"></span>${miner.id}: <span style="color: #dc3545; font-weight: bold;">Miner Unreachable</span></h4>`;
                        allPoolsHtml += '</div>'; // Close miner-card
                    } else {
                        const formattedHashrate = formatDeviceHashrate(miner.hashRate); // Use the specific device hashrate formatter.
                        const formattedExpected = formatDeviceHashrate(miner.expectedHashrate);
                        const AsicTemp = safeToFixed(Number(miner.temp),2);
                        const VRTemp = safeToFixed(Number(miner.vrTemp),2);
                        const displayAsicTemp = `<font color="${getLimitColor(AsicTemp, ASICTempMap)}"><b>${AsicTemp} &deg;C</b></font>`;
                        const displayVRTemp = `<font color="${getLimitColor(VRTemp, VRTempMap)}"><b>${VRTemp} &deg;C</b></font>`;
                        const displayFanSpeed = `<font color="${getLimitColor(miner.fanspeed, FanSpeedMap)}"><b>${miner.fanspeed} %</b></font>`;
                        const formattedUpTime = formatUptime(miner.uptimeSeconds);
                        // Create 5-column layout: Header | Label | Value | Label | Value
                        allPoolsHtml += `<h4><span class="status-indicator status-online" style="margin-right: 8px;"></span>${miner.id} <div class="line-graph-icon chart-button" data-instance-id="${miner.id}" title="View ${miner.id} Statistics"></div>`;
                        // Add restart and settings icons if settings are enabled
                        if(!disableSettings){
                            allPoolsHtml += ` <img src="/public/icon/icons8-rotate-right-64-white.png" class="restart-button restart-icon-hover" data-instance-id="${miner.id}" title="Restart Instance" style="width: 20px; height: 20px; margin-left: 8px; vertical-align: middle; cursor: pointer;">`;
                            allPoolsHtml += ` <img src="/public/icon/icons8-audio-65-white.png" class="settings-button settings-icon-hover" data-instance-id="${miner.id}" title="Edit Settings" style="width: 20px; height: 20px; margin-left: 8px; vertical-align: middle; cursor: pointer;">`;
                        }
                        // Add information icon (always visible)
                        allPoolsHtml += ` <img src="/public/icon/icons8-information-64-white.png" class="info-button info-icon-hover" data-instance-id="${miner.id}" title="View Detailed Information" style="width: 20px; height: 20px; margin-left: 8px; vertical-align: middle; cursor: pointer;">`;
                        allPoolsHtml += `</h4><div class="details-grid-five-columns">`;
                        // Hash row: Hash | Expected: | Value | Current: | Value
                        allPoolsHtml += `<div class="category-header">Hashrate</div><strong>Expected:</strong><span>${formattedExpected}</span><strong>Current:</strong><span>${formattedHashrate}</span>`;
                        // Difficulty row: Difficulty | Best: | Value | Session: | Value
                        allPoolsHtml += `<div class="category-header">Difficulty</div><strong>Best:</strong><span>${miner.bestDiff}</span><strong>Session:</strong><span>${miner.bestSessionDiff}</span>`;
                        // Pool row: Pool | Diff: | Value | Shares: | Value
                        allPoolsHtml += `<div class="category-header">Pool</div><strong>Diff:</strong><span>${miner.poolDifficulty}</span><strong>Shares:</strong><span>${miner.sharesAccepted}</span>`;
                        //Response Time and Shares Rejected Count
                        const formattedSharesRejected = formatFieldValue('sharesRejected', miner.sharesRejected, miner);
                        allPoolsHtml += `<div class="category-header">Status</div><strong>Response Time:</strong><span>${miner.responseTime} ms</span><strong>Shares Rejected:</strong><span>${formattedSharesRejected}</span>`;
                        // Temp row: Temp | ASIC: | Value | VR: | Value
                        allPoolsHtml += `<div class="category-header">Temperature</div><strong>ASIC:</strong><span>${displayAsicTemp}</span><strong>Voltage Regulator:</strong><span>${displayVRTemp}</span>`;
                        // Fan row: Fan | Speed: | Value | RPM: | Value
                        allPoolsHtml += `<div class="category-header">Fan</div><strong>Speed:</strong><span>${displayFanSpeed}</span><strong>RPM:</strong><span>${miner.fanrpm}</span>`;
                        //Uptime
                        allPoolsHtml += `<div class="category-header">General</div><strong>Frequency:</strong><span>${miner.frequency}</span><strong>Up Time:</strong><span>${formattedUpTime}</span>`;
                        allPoolsHtml += `<div class="category-header">Stratum</div><strong>Host:</strong><span>${miner.stratumURL}</span><strong>Port:</strong><span>${miner.stratumPort}</span>`;
                        allPoolsHtml += `</div>`; // Close details-grid-five-columns for individual miner status
                        allPoolsHtml += '</div>'; // Close miner-card
                    }
                });
            }

        allPoolsHtml += '</div>'; // Close miner-cards-container
        allPoolsHtml += `</div>`; // Close collapsible-content
        allPoolsHtml += `</div>`; // Close individual miner status card

        // Only show pool data if mining core is enabled
        if (miningCoreEnabled) {
            // Check if mining core data is available
            if (data && data.length > 0) {
                // Create single Mining Pool Status wrapper section
                allPoolsHtml += `<div class="mining-pool-status-section">`;
                allPoolsHtml += '<h3><span class="collapse-button" data-target="mining-pool-content">−</span> Mining Pool Status</h3>';
                allPoolsHtml += '<div id="mining-pool-content" class="collapsible-content">';
                allPoolsHtml += '<div class="pool-cards-container">'; // New container for responsive pool card layout

                // Loop through each mining core instance
                data.forEach((miningCoreInstance) => {
                    const instanceName = miningCoreInstance.instanceName;
                    const instanceStatus = miningCoreInstance.status;
                    const pools = miningCoreInstance.pools || [];

                    if (instanceStatus === 'Error') {
                        // Show error state for this instance as a card
                        allPoolsHtml += `<div class="pool-card">`;
                        allPoolsHtml += `<h4><span class="status-indicator status-error" style="margin-right: 8px;"></span>${instanceName}: <span style="color: #dc3545; font-weight: bold;">Mining Core Unreachable</span></h4>`;
                        allPoolsHtml += `<div class="details-grid">`;
                        allPoolsHtml += `<strong>Message:</strong> <span>${miningCoreInstance.message || 'Could not connect to mining core'}</span>`;
                        allPoolsHtml += `<strong>Note:</strong> <span>Mining core data is not available, but individual miners are still monitored.</span>`;
                        allPoolsHtml += `</div>`;
                        allPoolsHtml += `</div>`; // Close pool-card
                    } else if (pools.length > 0) {
                        // Show pools for this instance
                        pools.forEach((poolData) => { // Loop through each pool in this instance
                            allPoolsHtml += `<div class="pool-card">`; // Individual pool card wrapper
                            allPoolsHtml += `<h4><span class="status-indicator status-online" style="margin-right: 8px;"></span>${poolData.id.toUpperCase()} (${poolData.coin.symbol} - ${poolData.paymentProcessing.payoutScheme})</h4>`; // Pool specific heading

                            displayFields.forEach(categoryObj => {
                                const categoryName = Object.keys(categoryObj)[0];
                                const fieldsArray = categoryObj[categoryName];

                                // Create field data mapping for easier access
                                const fieldData = {};
                                fieldsArray.forEach(fieldObj => {
                                    let fieldKey = Object.keys(fieldObj)[0];
                                    const fieldLabel = fieldObj[fieldKey];

                                    // Correct the typo from the config file for 'lasNetworkBlockTime'.
                                    if (fieldKey === 'lasNetworkBlockTime') {
                                        fieldKey = 'lastNetworkBlockTime';
                                    }

                                    const displayValue = getNestedMiningCoreValue(fieldKey, poolData);
                                    const formattedValue = formatFieldValue(fieldKey, displayValue);
                                    fieldData[fieldKey] = { label: fieldLabel, value: formattedValue };
                                });

                                // Generate custom 5-column layouts based on category
                                if (categoryName === 'Network Status') {
                                    // Add instance name to Network Status header
                                    allPoolsHtml += `<h4>${categoryName} - ${instanceName}</h4><div class="details-grid-five-columns">`;
                                    // Network row
                                    allPoolsHtml += `<div class="category-header">Network</div><strong>Difficulty:</strong><span>${fieldData.networkDifficulty?.value || 'N/A'}</span><strong>Hashrate:</strong><span>${fieldData.networkHashrate?.value || 'N/A'}</span>`;
                                    // Block row
                                    allPoolsHtml += `<div class="category-header">Block</div><strong>Height:</strong><span>${fieldData.blockHeight?.value || 'N/A'}</span><strong>Last Block Time:</strong><span>${fieldData.lastNetworkBlockTime?.value || 'N/A'}</span>`;
                                    // General row
                                    allPoolsHtml += `<div class="category-header">General</div><strong>Connected Peers:</strong><span>${fieldData.connectedPeers?.value || 'N/A'}</span><strong>Node Version:</strong><span>${fieldData.nodeVersion?.value || 'N/A'}</span>`;
                                    allPoolsHtml += `</div>`;
                                } else if (categoryName === 'Miner(s) Status') {
                                    allPoolsHtml += `<h4>${categoryName}</h4><div class="details-grid-five-columns">`;
                                    // Status row
                                    allPoolsHtml += `<div class="category-header">Status</div><strong>Connected Miners:</strong><span>${fieldData.connectedMiners?.value || 'N/A'}</span><strong>Pool Hashrate:</strong><span>${fieldData.poolHashrate?.value || 'N/A'}</span>`;
                                    allPoolsHtml += `</div>`;
                                } else if (categoryName === 'Rewards' || categoryName === 'Rewards Status') {
                                    allPoolsHtml += `<h4>${categoryName}</h4><div class="details-grid-five-columns">`;
                                    // Total row (Paid and Blocks)
                                    allPoolsHtml += `<div class="category-header">Total</div><strong>Paid:</strong><span>${fieldData.totalPaid?.value || 'N/A'}</span><strong>Blocks:</strong><span>${fieldData.totalBlocks?.value || 'N/A'}</span>`;
                                    // Total row (Confirmed and Pending Blocks)
                                    allPoolsHtml += `<div class="category-header">Total</div><strong>Confirmed Blocks:</strong><span>${fieldData.totalConfirmedBlocks?.value || 'N/A'}</span><strong>Pending Blocks:</strong><span>${fieldData.totalPendingBlocks?.value || 'N/A'}</span>`;
                                    // Reward row
                                    allPoolsHtml += `<div class="category-header">Reward</div><strong>Block Reward:</strong><span>${fieldData.blockReward?.value || 'N/A'}</span><strong>Pool Block Time:</strong><span>${fieldData.lastPoolBlockTime?.value || 'N/A'}</span>`;
                                    allPoolsHtml += `</div>`;
                                } else {
                                    // Fallback to original layout for unknown categories
                                    allPoolsHtml += `<h4>${categoryName}</h4><div class="details-grid">`;
                                    fieldsArray.forEach(fieldObj => {
                                        let fieldKey = Object.keys(fieldObj)[0];
                                        const fieldLabel = fieldObj[fieldKey];

                                        if (fieldKey === 'lasNetworkBlockTime') {
                                            fieldKey = 'lastNetworkBlockTime';
                                        }

                                        const displayValue = getNestedMiningCoreValue(fieldKey, poolData);
                                        const formattedValue = formatFieldValue(fieldKey, displayValue);

                                        allPoolsHtml += `<strong>${fieldLabel}:</strong> <span>${formattedValue}</span>`;
                                    });
                                    allPoolsHtml += `</div>`;
                                }
                            });
                            allPoolsHtml += `</div>`; // Close pool-card
                        });
                    } else if (instanceStatus === 'OK') {
                        // No pools available for this instance but it's reachable
                        allPoolsHtml += `<div class="pool-card">`;
                        allPoolsHtml += `<h4><span class="status-indicator status-online" style="margin-right: 8px;"></span>${instanceName}</h4>`;
                        allPoolsHtml += `<div class="details-grid">`;
                        allPoolsHtml += `<strong>Note:</strong> <span>No pools configured for this mining core instance.</span>`;
                        allPoolsHtml += `</div>`;
                        allPoolsHtml += `</div>`; // Close pool-card
                    }
                });

                allPoolsHtml += '</div>'; // Close pool-cards-container
                allPoolsHtml += `</div>`; // Close collapsible-content
                allPoolsHtml += `</div>`; // Close mining-pool-status-section
            } else {
                // Show message when no mining core instances are configured
                allPoolsHtml += `<div class="mining-pool-status-section">`;
                allPoolsHtml += `<h3><span class="collapse-button" data-target="mining-pool-content">−</span> Mining Pool Status</h3>`;
                allPoolsHtml += '<div id="mining-pool-content" class="collapsible-content">';
                allPoolsHtml += `<h4><span class="status-indicator status-error" style="margin-right: 8px;"></span><span style="color: #dc3545; font-weight: bold;">No Mining Core Instances Configured</span></h4>`;
                allPoolsHtml += `<div class="details-grid">`;
                allPoolsHtml += `<strong>Note:</strong> <span>Mining core is enabled but no instances are configured.</span>`;
                allPoolsHtml += `</div>`;
                allPoolsHtml += `</div>`;
                allPoolsHtml += `</div>`;
            }
        }

        // Add Crypto Node Status section
        allPoolsHtml += generateCryptoNodeStatusHtml(cryptoNodeData);

        return allPoolsHtml;
    }

    /**
     * Generates the HTML for the Crypto Node Status section
     * @param {Array} cryptoNodes - Array of crypto node data objects
     * @returns {string} The HTML string for the crypto node status
     */
    function generateCryptoNodeStatusHtml(cryptoNodes) {
        let html = '';

        // Only render if crypto node data exists and is not empty
        if (!cryptoNodes || cryptoNodes.length === 0) {
            return html;
        }

        // Create Crypto Node Status wrapper section
        html += `<div class="crypto-node-status-section">`;
        html += '<h3><span class="collapse-button" data-target="crypto-node-content">−</span> Crypto Node Status</h3>';
        html += '<div id="crypto-node-content" class="collapsible-content">';
        html += '<div class="crypto-node-cards-container">'; // Container for responsive node card layout

        cryptoNodes.forEach((nodeData) => {
            html += `<div class="crypto-node-card">`; // Individual node card wrapper

            if (nodeData.status === 'Error') {
                // Display error state
                html += `<h4><span class="status-indicator status-error" style="margin-right: 8px;"></span>${nodeData.id}: <span style="color: #dc3545; font-weight: bold;">Node Unreachable</span></h4>`;
                html += `<div class="details-grid">`;
                html += `<strong>Status:</strong> <span style="color: #dc3545;">Error</span>`;
                html += `<strong>Message:</strong> <span>${nodeData.message || 'Could not connect to node'}</span>`;
                html += `</div>`;
            } else {
                // Display node name with online indicator and algorithm
                const algoText = nodeData.nodeAlgo ? ` - ${nodeData.nodeAlgo}` : '';
                html += `<h4><span class="status-indicator status-online" style="margin-right: 8px;"></span>${nodeData.id} (${nodeData.nodeType.toUpperCase()}${algoText})</h4>`;

                // Render all display fields in a single 10-column grid
                if (nodeData.displayFields && Array.isArray(nodeData.displayFields)) {
                    html += `<div class="details-grid-ten-columns">`;

                    nodeData.displayFields.forEach(categoryObj => {
                        const categoryName = Object.keys(categoryObj)[0];
                        const fieldsArray = categoryObj[categoryName];

                        // Add category header as h4 (spans full width)
                        html += `<h4>${categoryName}</h4>`;

                        // Process fields in groups of 5 for the ten-column layout (5 label/value pairs = 10 columns)
                        for (let i = 0; i < fieldsArray.length; i += 5) {
                            // Add up to 5 fields per row
                            for (let j = 0; j < 5; j++) {
                                const field = fieldsArray[i + j];

                                if (field) {
                                    const fieldKey = Object.keys(field)[0];
                                    const fieldLabel = field[fieldKey];
                                    const displayValue = getCryptoNodeValue(fieldKey, nodeData);
                                    const formattedValue = formatCryptoNodeValue(fieldKey, displayValue);

                                    html += `<strong>${fieldLabel}:</strong><span>${formattedValue}</span>`;
                                } else {
                                    // Fill empty cells if we don't have 5 fields
                                    html += `<div></div><div></div>`;
                                }
                            }
                        }
                    });

                    html += `</div>`; // Close details-grid-ten-columns
                }
            }

            html += `</div>`; // Close crypto-node-card
        });

        html += '</div>'; // Close crypto-node-cards-container
        html += `</div>`; // Close collapsible-content
        html += `</div>`; // Close crypto-node-status-section

        return html;
    }

    /**
     * Retrieves a value from crypto node data based on field key
     * @param {string} fieldKey - The field key to look up (supports nested paths with '/' separator)
     * @param {object} nodeData - The crypto node data object
     * @returns {*} The value or 'N/A' if not found
     */
    function getCryptoNodeValue(fieldKey, nodeData) {
        // Handle nested paths (e.g., "difficulties/sha256d")
        if (fieldKey.includes('/')) {
            const parts = fieldKey.split('/');
            const parentKey = parts[0];
            const childKey = parts[1];

            // Check blockchainInfo for nested values
            if (nodeData.blockchainInfo && parentKey in nodeData.blockchainInfo) {
                const parentValue = nodeData.blockchainInfo[parentKey];
                if (parentValue && typeof parentValue === 'object' && childKey in parentValue) {
                    return parentValue[childKey];
                }
            }

            // Check networkTotals for nested values
            if (nodeData.networkTotals && parentKey in nodeData.networkTotals) {
                const parentValue = nodeData.networkTotals[parentKey];
                if (parentValue && typeof parentValue === 'object' && childKey in parentValue) {
                    return parentValue[childKey];
                }
            }

            // Check networkInfo for nested values
            if (nodeData.networkInfo && parentKey in nodeData.networkInfo) {
                const parentValue = nodeData.networkInfo[parentKey];
                if (parentValue && typeof parentValue === 'object' && childKey in parentValue) {
                    return parentValue[childKey];
                }
            }

            return 'N/A';
        }

        // Check blockchainInfo
        if (nodeData.blockchainInfo && fieldKey in nodeData.blockchainInfo) {
            return nodeData.blockchainInfo[fieldKey];
        }

        // Check networkInfo
        if (nodeData.networkInfo && fieldKey in nodeData.networkInfo) {
            return nodeData.networkInfo[fieldKey];
        }

        // Check networkTotals
        if (nodeData.networkTotals && fieldKey in nodeData.networkTotals) {
            return nodeData.networkTotals[fieldKey];
        }

        // Check uploadtarget (nested in networkTotals)
        if (nodeData.networkTotals && nodeData.networkTotals.uploadtarget && fieldKey in nodeData.networkTotals.uploadtarget) {
            return nodeData.networkTotals.uploadtarget[fieldKey];
        }

        // Check balance (direct value)
        if (fieldKey === 'balance' && nodeData.balance !== undefined) {
            return nodeData.balance;
        }

        return 'N/A';
    }

    /**
     * Formats crypto node values for display
     * @param {string} fieldKey - The field key
     * @param {*} value - The raw value
     * @returns {string} Formatted value string
     */
    function formatCryptoNodeValue(fieldKey, value) {
        if (value === null || value === undefined || value === 'N/A') {
            return 'N/A';
        }

        // Handle nested paths for formatting (e.g., "difficulties/sha256d")
        if (fieldKey.includes('/')) {
            const parts = fieldKey.split('/');
            const childKey = parts[1];

            // Format difficulty values
            if (parts[0] === 'difficulties' || childKey.includes('difficulty')) {
                return formatLargeNumber(value);
            }
        }

        // Format based on field type
        switch (fieldKey) {
            case 'balance':
                return safeToFixed(value, 8) + ' DGB';

            case 'sha256d':
            case 'difficulty':
                return formatLargeNumber(value);

            case 'size_on_disk':
            case 'totalbytesrecv':
            case 'totalbytessent':
            case 'bytes_left_in_cycle':
                return formatBytes(value);

            case 'verificationprogress':
                return (value * 100).toFixed(2) + '%';

            case 'mediantime':
                return new Date(value * 1000).toLocaleString();

            case 'timemillis':
                return new Date(value).toLocaleString();

            case 'time_left_in_cycle':
            case 'timeframe':
                return formatSeconds(value);

            case 'initialblockdownload':
            case 'pruned':
            case 'networkactive':
            case 'target_reached':
            case 'serve_historical_blocks':
                return value ? 'Yes' : 'No';

            case 'warnings':
                if (typeof value === 'string' && value.trim() === '') {
                    return 'None';
                }
                return value || 'None';
            case 'target':
                return formatBytes(value);
                
            default:
                // Return as-is for other fields
                if (typeof value === 'object') {
                    return JSON.stringify(value);
                }
                return String(value);
        }
    }

    /**
     * Formats bytes to human-readable format
     * @param {number} bytes - Number of bytes
     * @returns {string} Formatted byte string
     */
    function formatBytes(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    /**
     * Formats seconds to human-readable time
     * @param {number} seconds - Number of seconds
     * @returns {string} Formatted time string
     */
    function formatSeconds(seconds) {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;
        return `${hours}h ${minutes}m ${secs}s`;
    }

    /**
     * Generates the detailed HTML for a single device in the right pane using display_fields.
     * @param {object} data - The device data object.
     * @returns {string} The HTML string for the device details.
     */
    function generateDeviceDetailsHtml(data) {
        // Display a detailed error message if data fetching failed for this instance.
        if (data.status === 'Error') {
            return `
                <h2 style="color: #d8000c;">Device: ${data.hostname || 'Unknown Device'} - Error</h2>
                <div class="details-grid">
                    <strong>Status:</strong> <span style="font-weight: bold; color: #d8000c;">Error</span>
                    <strong>Message:</strong> <span>${data.message || 'Could not retrieve data for this device.'}</span>
                    <strong>Suggestion:</strong> <span>Please check the device's network connection and API status, then refresh the dashboard.</span>
                </div>
            `;
        }

        let html = `<h2>${ data.id || 'Unknown Device'}</h2>`;

        displayFieldsConfig.forEach(categoryObj => {
            const categoryName = Object.keys(categoryObj)[0];
            const fieldsArray = categoryObj[categoryName];

            html += `<h3>${categoryName}</h3><div class="details-grid">`;
            fieldsArray.forEach(fieldObj => {
                const fieldKey = Object.keys(fieldObj)[0];
                const fieldLabel = fieldObj[fieldKey];
                const displayValue = data[fieldKey];
                const formattedValue = formatFieldValue(fieldKey, displayValue, data);

                html += `<strong>${fieldLabel}:</strong> <span>${formattedValue}</span>`;
            });
            html += `</div>`;
        });

        return html;
    }

    // --- UI Population and Event Handling ---

    /**
     * Adds the logout button to the header if authentication is enabled.
     */
    function addLogoutButton() {
        // Only add the logout button if authentication is enabled
        if (!disableAuthentication) {
            const header = document.querySelector('header');
            
            if (header) {
                // Check if button already exists to avoid duplicates
                const existingButton = document.getElementById('logout-button');
                if (!existingButton) {
                    const logoutButton = document.createElement('div');
                    logoutButton.id = 'logout-button';
                    logoutButton.title = 'Logout from dashboard';
                    header.appendChild(logoutButton);

                    logoutButton.addEventListener('click', async () => {
                        try {
                            const response = await fetch('/api/logout', {
                                method: 'POST'
                            });
                            if (response.ok) {
                                // On successful logout, the server clears the session cookie.
                                // Redirect the user to the login page.
                                window.location.href = '/login';
                            } else {
                                const result = await response.json();
                                alert(`Logout failed: ${result.message || 'Unknown error'}`);
                            }
                        } catch (error) {
                            console.error('Logout request failed:', error);
                            alert('Failed to send logout command. See console for details.');
                        }
                    });
                }
            }
        }
    }


    /**
     * Displays the mining dashboard content (formerly the summary view)
     */
    function displayDashboard() {
        // Hide/show configuration icon based on disable_configurations setting
        if (configIcon) {
            configIcon.style.display = disableConfigurations ? 'none' : 'inline-block';
        }

        // Display the mining core summary directly in the dashboard
        miningCoreDetailsDiv.innerHTML = generateMiningCoreDetailsHtml(miningCoreData, miningCoreDisplayFields);

        // Add event listeners to Chart buttons
        attachChartButtonEventListeners();

        // Add event listeners to Restart and Settings buttons
        attachRestartAndSettingsButtonEventListeners();

        // Add event listeners to Collapse buttons
        attachCollapseButtonEventListeners();

        // Restore saved section states
        restoreSavedSectionStates();
    }

    /**
     * Checks if a configuration migration was performed and shows a notification modal
     */
    async function checkConfigurationMigration() {
        try {
            const response = await fetch('/api/migration/status');
            const result = await response.json();

            if (result.success && result.data && result.data.migrated) {
                showMigrationNotification(result.data);
            }
        } catch (error) {
            console.error('Error checking migration status:', error);
        }
    }

    /**
     * Shows a modal notification about the configuration migration
     * @param {object} migrationData - Migration status data
     */
    function showMigrationNotification(migrationData) {
        const existingModal = document.getElementById('migration-modal');
        if (existingModal) existingModal.remove();

        let migrationsHtml = '<ul style="text-align: left; margin: 10px 0;">';
        migrationData.migrations.forEach(migration => {
            migrationsHtml += `<li>${migration}</li>`;
        });
        migrationsHtml += '</ul>';

        const modalHtml = `
            <div id="migration-modal" class="modal">
                <div class="modal-content" style="max-width: 600px;">
                    <h2 style="color: #4CAF50;">Configuration Updated</h2>
                    <p style="font-size: 1.1em; margin: 20px 0;">
                        Your configuration has been automatically migrated to the latest format.
                    </p>
                    <p style="margin: 10px 0;">
                        <strong>Changes applied:</strong>
                    </p>
                    ${migrationsHtml}
                    <p style="margin: 20px 0; font-size: 0.95em; color: #aaa;">
                        No action is required. Your settings have been preserved.
                    </p>
                    <div class="modal-actions">
                        <button type="button" class="animated-button confirm-button" id="migration-ok-btn">OK</button>
                    </div>
                </div>
            </div>`;

        document.body.insertAdjacentHTML('beforeend', modalHtml);

        const modal = document.getElementById('migration-modal');
        const okButton = document.getElementById('migration-ok-btn');

        const closeModal = async () => {
            // Clear the migration status on the backend
            try {
                await fetch('/api/migration/clear', { method: 'POST' });
            } catch (error) {
                console.error('Error clearing migration status:', error);
            }
            modal.remove();
        };

        okButton.addEventListener('click', closeModal);
        window.addEventListener('click', (event) => {
            if (event.target === modal) closeModal();
        });
    }

    // Initialize the dashboard
    displayDashboard();

});
