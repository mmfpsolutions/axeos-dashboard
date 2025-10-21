/**
 * @file Statistics Modal Service
 * Handles real-time statistics visualization for individual BitAxe miners
 * Features Chart.js integration with dual Y-axis charts (lazy-loaded)
 */

const statisticsModal = (() => {
    let currentChart = null;
    let pollInterval = null;
    let currentInstanceId = null;
    let chartJsLoaded = false;
    let chartJsLoading = false;

    /**
     * Formats hashrate with appropriate units (MH/s, GH/s, TH/s)
     * @param {number} hashrate - Hashrate in GH/s
     * @returns {string} Formatted hashrate string
     */
    function formatHashrate(hashrate) {
        if (typeof hashrate !== 'number' || isNaN(hashrate) || hashrate < 0) {
            return 'N/A';
        }

        if (hashrate < 1) {
            return `${(hashrate * 1000).toFixed(1)} Mh/s`;
        } else if (hashrate < 1000) {
            return `${hashrate.toFixed(1)} Gh/s`;
        } else {
            return `${(hashrate / 1000).toFixed(2)} Th/s`;
        }
    }

    /**
     * Lazy-loads Chart.js library if not already loaded
     * @returns {Promise} Resolves when Chart.js is loaded
     */
    function loadChartJs() {
        if (chartJsLoaded) {
            return Promise.resolve();
        }

        if (chartJsLoading) {
            // Wait for existing load to complete
            return new Promise((resolve) => {
                const checkLoaded = setInterval(() => {
                    if (chartJsLoaded) {
                        clearInterval(checkLoaded);
                        resolve();
                    }
                }, 100);
            });
        }

        chartJsLoading = true;

        return new Promise((resolve, reject) => {
            // Load Chart.js
            const chartScript = document.createElement('script');
            chartScript.src = 'https://cdn.jsdelivr.net/npm/chart.js';
            chartScript.onload = () => {
                // Load date adapter
                const adapterScript = document.createElement('script');
                adapterScript.src = 'https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns';
                adapterScript.onload = () => {
                    chartJsLoaded = true;
                    chartJsLoading = false;
                    resolve();
                };
                adapterScript.onerror = () => {
                    chartJsLoading = false;
                    reject(new Error('Failed to load Chart.js date adapter'));
                };
                document.head.appendChild(adapterScript);
            };
            chartScript.onerror = () => {
                chartJsLoading = false;
                reject(new Error('Failed to load Chart.js'));
            };
            document.head.appendChild(chartScript);
        });
    }

    /**
     * Creates and displays the statistics modal for a specific miner instance
     * @param {string} instanceId - The ID of the miner instance to show statistics for
     */
    async function openStatisticsModal(instanceId) {
        const existingModal = document.getElementById('statistics-modal');
        if (existingModal) existingModal.remove();

        currentInstanceId = instanceId;

        const modalHtml = generateModalHtml(instanceId);
        document.body.insertAdjacentHTML('beforeend', modalHtml);

        const modal = document.getElementById('statistics-modal');
        const closeModal = () => {
            stopPolling();
            if (currentChart) {
                currentChart.destroy();
                currentChart = null;
            }
            modal.remove();
        };

        // Event listeners
        modal.querySelector('.close-button').addEventListener('click', closeModal);
        window.addEventListener('click', (event) => {
            if (event.target === modal) closeModal();
        });

        // Show loading message for chart library
        const loadingDiv = modal.querySelector('#stats-loading');
        const loadingSpan = loadingDiv.querySelector('span');
        loadingSpan.textContent = 'Loading chart library...';

        try {
            // Lazy-load Chart.js before initializing
            await loadChartJs();

            // Initialize chart and start data polling
            initializeChart();
            startPolling();
        } catch (error) {
            console.error('Failed to load Chart.js:', error);
            const errorDiv = modal.querySelector('#stats-error');
            const errorMessage = modal.querySelector('#error-message');
            loadingDiv.style.display = 'none';
            errorDiv.style.display = 'flex';
            errorMessage.textContent = error.message;
        }
    }

    /**
     * Generates the HTML structure for the statistics modal
     * @param {string} instanceId - The miner instance ID
     * @returns {string} The complete HTML for the modal
     */
    function generateModalHtml(instanceId) {
        return `
            <div id="statistics-modal" class="modal">
                <div class="modal-content">
                    <div class="statistics-modal-header">
                        <h2>${instanceId} - Real-time Statistics</h2>
                        <span class="close-button">&times;</span>
                    </div>
                    <div class="statistics-modal-body">
                        <div class="stats-status-info">
                            <div class="stats-status-left">
                                <div class="stats-status-item">
                                    <span id="stats-status-indicator" class="stats-status-indicator"></span>
                                    <span id="stats-status-text">Connecting...</span>
                                </div>
                                <div class="stats-status-item">
                                    <span>Update interval: 25s</span>
                                </div>
                                <div class="stats-status-item">
                                    <span>Hashrate: <span id="stats-current-hashrate">--</span></span>
                                </div>
                                <div class="stats-status-item">
                                    <span>Average: <span id="stats-hashrate-average">--</span></span>
                                </div>
                                <div class="stats-status-item">
                                    <span>ASIC Temp: <span id="stats-current-temperature">--</span></span>
                                </div>
                            </div>
                            <div class="stats-last-update" id="stats-last-update">
                                Initializing...
                            </div>
                        </div>
                        <div class="stats-chart-legend">
                            <div class="stats-legend-item">
                                <div class="stats-legend-color hashrate"></div>
                                <span>Hashrate</span>
                            </div>
                            <div class="stats-legend-item">
                                <div class="stats-legend-color temperature"></div>
                                <span>ASIC Temp</span>
                            </div>
                        </div>
                        <div class="stats-chart-container">
                            <div id="stats-loading" class="stats-loading">
                                <div class="stats-loading-spinner"></div>
                                <span>Loading statistics data...</span>
                            </div>
                            <div id="stats-error" class="stats-error" style="display: none;">
                                <div class="stats-error-icon">⚠</div>
                                <div>Failed to load statistics data</div>
                                <div style="font-size: 0.9em; margin-top: 5px;" id="error-message"></div>
                            </div>
                            <div class="stats-chart-wrapper" style="display: none;">
                                <canvas id="statistics-chart" class="stats-chart"></canvas>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    /**
     * Initializes the Chart.js chart with dual Y-axis configuration
     */
    function initializeChart() {
        const canvas = document.getElementById('statistics-chart');
        const ctx = canvas.getContext('2d');

        currentChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [
                    {
                        label: 'Hashrate',
                        data: [],
                        backgroundColor: 'rgba(255, 23, 68, 0.2)',
                        borderColor: '#FF1744',
                        borderWidth: 1.5,
                        fill: true,
                        pointRadius: 2,
                        pointBackgroundColor: 'transparent',
                        pointBorderColor: '#FF1744',
                        pointBorderWidth: 2,
                        tension: 0.1,
                        yAxisID: 'y'
                    },
                    {
                        label: 'ASIC Temperature',
                        data: [],
                        backgroundColor: 'transparent',
                        borderColor: '#f5f5f5',
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 2,
                        pointBackgroundColor: 'transparent',
                        pointBorderColor: '#f5f5f5',
                        pointBorderWidth: 2,
                        tension: 0.1,
                        yAxisID: 'y1'
                    }
                ]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                animation: {
                    duration: 0 // Disable animations for real-time feel
                },
                interaction: {
                    mode: 'index',
                    intersect: false,
                },
                plugins: {
                    legend: {
                        display: false // We have custom legend
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false,
                        callbacks: {
                            label: function(context) {
                                let label = context.dataset.label || '';
                                if (label) {
                                    label += ': ';
                                }
                                if (context.datasetIndex === 0) {
                                    // Hashrate
                                    label += formatHashrate(context.parsed.y);
                                } else {
                                    // Temperature
                                    label += parseFloat(context.parsed.y).toFixed(1) + ' °C';
                                }
                                return label;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        type: 'time',
                        time: {
                            unit: 'hour',
                            displayFormats: {
                                hour: 'HH:mm'
                            }
                        },
                        title: {
                            display: true,
                            text: 'Time',
                            color: '#666'
                        },
                        grid: {
                            color: 'rgba(200, 200, 200, 0.3)'
                        }
                    },
                    y: {
                        type: 'linear',
                        display: true,
                        position: 'left',
                        title: {
                            display: true,
                            text: 'Hashrate',
                            color: '#FF1744'
                        },
                        ticks: {
                            color: '#FF1744',
                            callback: function(value) {
                                return formatHashrate(value);
                            }
                        },
                        grid: {
                            color: 'rgba(255, 23, 68, 0.1)'
                        }
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                            display: true,
                            text: 'ASIC Temp',
                            color: '#f5f5f5'
                        },
                        ticks: {
                            color: '#f5f5f5',
                            callback: function(value) {
                                return parseFloat(value).toFixed(1) + '°C';
                            }
                        },
                        grid: {
                            drawOnChartArea: false,
                            color: 'rgba(245, 245, 245, 0.3)'
                        },
                        suggestedMin: 50,
                        suggestedMax: 80
                    }
                }
            }
        });
    }

    /**
     * Fetches statistics data from the backend API
     * @returns {Promise<Object>} The statistics data
     */
    async function fetchStatisticsData() {
        try {
            const response = await fetch(`/api/statistics?instanceId=${encodeURIComponent(currentInstanceId)}`);
            
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ message: 'Unknown error' }));
                throw new Error(errorData.message || `HTTP ${response.status}`);
            }

            const result = await response.json();
            
            if (!result.success) {
                throw new Error(result.message || 'API request failed');
            }

            return result.data;
        } catch (error) {
            console.error('Failed to fetch statistics:', error);
            throw error;
        }
    }

    /**
     * Processes and adds new data points to the chart
     * @param {Object} statisticsData - The raw statistics data from the API
     */
    function updateChart(statisticsData) {
        if (!currentChart || !statisticsData || !statisticsData.statistics) {
            return;
        }

        const statistics = statisticsData.statistics;
        const currentTimestamp = statisticsData.currentTimestamp;

        // Clear existing data and rebuild from scratch for consistency
        currentChart.data.labels = [];
        currentChart.data.datasets[0].data = [];
        currentChart.data.datasets[1].data = [];

        // Process each data point
        statistics.forEach(dataPoint => {
            if (dataPoint && dataPoint.length >= 4) {
                const hashrate = dataPoint[0]; // GH/s
                const temperature = dataPoint[1]; // °C  
                const power = dataPoint[2]; // Watts (unused for now)
                const timeOffset = dataPoint[3]; // milliseconds offset

                // Calculate actual timestamp
                const actualTimestamp = new Date().getTime() - currentTimestamp + timeOffset;
                
                currentChart.data.labels.push(actualTimestamp);
                currentChart.data.datasets[0].data.push(hashrate);
                currentChart.data.datasets[1].data.push(temperature);
            }
        });

        // Limit data points to prevent performance issues (keep last 500 points)
        const maxPoints = 500;
        if (currentChart.data.labels.length > maxPoints) {
            const excess = currentChart.data.labels.length - maxPoints;
            currentChart.data.labels.splice(0, excess);
            currentChart.data.datasets[0].data.splice(0, excess);
            currentChart.data.datasets[1].data.splice(0, excess);
        }

        // Update current stats display
        updateCurrentStats(statistics);
        
        currentChart.update('none'); // Update without animation
        updateStatusIndicators(true);
    }

    /**
     * Updates the current statistics display with the latest values
     * @param {Array} statistics - Array of statistics data points
     */
    function updateCurrentStats(statistics) {
        const currentHashrateElement = document.getElementById('stats-current-hashrate');
        const hashrateAverageElement = document.getElementById('stats-hashrate-average');
        const currentTemperatureElement = document.getElementById('stats-current-temperature');

        if (!statistics || statistics.length === 0) {
            if (currentHashrateElement) currentHashrateElement.textContent = '--';
            if (hashrateAverageElement) hashrateAverageElement.textContent = '--';
            if (currentTemperatureElement) currentTemperatureElement.textContent = '--';
            return;
        }

        // Get current values (last data point)
        const lastDataPoint = statistics[statistics.length - 1];
        if (lastDataPoint && lastDataPoint.length >= 2) {
            const currentHashrate = lastDataPoint[0]; // GH/s
            const currentTemperature = lastDataPoint[1]; // °C

            // Calculate average hashrate from all available data points
            let totalHashrate = 0;
            let validPoints = 0;
            statistics.forEach(dataPoint => {
                if (dataPoint && dataPoint.length >= 1 && typeof dataPoint[0] === 'number') {
                    totalHashrate += dataPoint[0];
                    validPoints++;
                }
            });
            const averageHashrate = validPoints > 0 ? totalHashrate / validPoints : 0;

            // Update the display elements
            if (currentHashrateElement) {
                currentHashrateElement.textContent = formatHashrate(currentHashrate);
            }
            if (hashrateAverageElement) {
                hashrateAverageElement.textContent = formatHashrate(averageHashrate);
            }
            if (currentTemperatureElement) {
                currentTemperatureElement.textContent = `${currentTemperature.toFixed(1)}°C`;
            }
        }
    }

    /**
     * Updates the status indicators and timestamps
     * @param {boolean} success - Whether the last update was successful
     * @param {string} errorMessage - Error message if applicable
     */
    function updateStatusIndicators(success, errorMessage = null) {
        const statusIndicator = document.getElementById('stats-status-indicator');
        const statusText = document.getElementById('stats-status-text');
        const lastUpdate = document.getElementById('stats-last-update');

        if (success) {
            statusIndicator.className = 'stats-status-indicator';
            statusText.textContent = 'Live';
            lastUpdate.textContent = `Last update: ${new Date().toLocaleTimeString()}`;
        } else {
            statusIndicator.className = 'stats-status-indicator error';
            statusText.textContent = 'Error';
            lastUpdate.textContent = errorMessage || 'Connection failed';
        }
    }

    /**
     * Shows error state in the modal
     * @param {string} message - Error message to display
     */
    function showError(message) {
        const loading = document.getElementById('stats-loading');
        const error = document.getElementById('stats-error');
        const chartWrapper = document.querySelector('.stats-chart-wrapper');
        const errorMessage = document.getElementById('error-message');

        loading.style.display = 'none';
        chartWrapper.style.display = 'none';
        error.style.display = 'flex';
        errorMessage.textContent = message;

        updateStatusIndicators(false, message);
    }

    /**
     * Shows the chart and hides loading/error states
     */
    function showChart() {
        const loading = document.getElementById('stats-loading');
        const error = document.getElementById('stats-error');
        const chartWrapper = document.querySelector('.stats-chart-wrapper');

        loading.style.display = 'none';
        error.style.display = 'none';
        chartWrapper.style.display = 'block';
    }

    /**
     * Starts polling for statistics data
     */
    function startPolling() {
        // Initial fetch
        fetchAndUpdate();

        // Set up polling every 25 seconds
        pollInterval = setInterval(fetchAndUpdate, 25000);
    }

    /**
     * Stops polling for statistics data
     */
    function stopPolling() {
        if (pollInterval) {
            clearInterval(pollInterval);
            pollInterval = null;
        }
    }

    /**
     * Fetches data and updates the chart
     */
    async function fetchAndUpdate() {
        try {
            const data = await fetchStatisticsData();
            showChart();
            updateChart(data);
        } catch (error) {
            console.error('Error updating statistics:', error);
            showError(error.message);
        }
    }

    // Public API
    return {
        openStatisticsModal
    };
})();

// Make it globally available
window.statisticsModal = statisticsModal;