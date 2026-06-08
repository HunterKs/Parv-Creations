/* Notification System (Toast Notifications) */

const NOTIFICATION_TYPES = {
    success: 'success',
    error: 'error',
    warning: 'warning',
    info: 'info'
};

class NotificationManager {
    constructor() {
        this.container = null;
        this.notifications = [];
        this.maxNotifications = 5;
        this.init();
    }

    init() {
        // Create notification container if it doesn't exist
        if (!document.getElementById('notification-container')) {
            this.container = document.createElement('div');
            this.container.id = 'notification-container';
            this.container.style.position = 'fixed';
            this.container.style.top = '1rem';
            this.container.style.right = '1rem';
            this.container.style.zIndex = '1000';
            this.container.style.display = 'flex';
            this.container.style.flexDirection = 'column';
            this.container.style.gap = '0.5rem';
            document.body.appendChild(this.container);
        } else {
            this.container = document.getElementById('notification-container');
        }
    }

    /**
     * Show a notification toast
     * @param {string} message - The message to display
     * @param {string} type - The type of notification (success, error, warning, info)
     * @param {number} duration - Duration in milliseconds (default: 5000)
     */
    show(message, type = NOTIFICATION_TYPES.info, duration = 5000) {
        // Remove oldest notification if we've reached the limit
        if (this.notifications.length >= this.maxNotifications) {
            const oldest = this.notifications.shift();
            if (oldest.element) {
                oldest.element.remove();
            }
        }

        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.display = 'flex';
        notification.style.alignItems = 'center';
        notification.style.padding = '1rem';
        notification.style.borderRadius = 'var(--border-radius)';
        notification.style.boxShadow = 'var(--box-shadow-lg)';
        notification.style.backgroundColor = 'white';
        notification.style.color = 'var(--gray-900)';
        notification.style.minWidth = '300px';
        notification.style.maxWidth = '400px';
        notification.style.position = 'relative';
        notification.style.animation = 'slideIn 0.3s ease-out forwards, fadeOut 0.3s ease-in ${duration - 300}ms forwards';

        // Define keyframes if not already defined
        if (!document.getElementById('notification-keyframes')) {
            const style = document.createElement('style');
            style.id = 'notification-keyframes';
            style.textContent = `
                @keyframes slideIn {
                    from { transform: translateX(100%); opacity: 0; }
                    to { transform: translateX(0); opacity: 1; }
                }
                @keyframes fadeOut {
                    to { transform: translateX(0); opacity: 0; }
                }
            `;
            document.head.appendChild(style);
        }

        // Set background color based on type
        let bgColor = 'var(--gray-100)';
        let textColor = 'var(--gray-900)';
        let iconColor = 'var(--primary-color)';

        switch (type) {
            case NOTIFICATION_TYPES.success:
                bgColor = '#dcfce7';
                textColor = '#166534';
                iconColor = '#10b981';
                break;
            case NOTIFICATION_TYPES.error:
                bgColor = '#fee2e2';
                textColor = '#991b1b';
                iconColor = '#ef4444';
                break;
            case NOTIFICATION_TYPES.warning:
                bgColor = '#fffbeb';
                textColor = '#92400e';
                iconColor = '#f59e0b';
                break;
            case NOTIFICATION_TYPES.info:
                bgColor = '#dbeafe';
                textColor = '#1e40af';
                iconColor = '#3b82f6';
                break;
        }

        // Apply dark mode adjustments
        if (document.documentElement.getAttribute('data-theme') === 'dark') {
            // In dark mode, we'll use different colors
            bgColor = 'var(--gray-800)';
            textColor = 'var(--gray-100)';
            switch (type) {
                case NOTIFICATION_TYPES.success:
                    bgColor = '#064e3b';
                    textColor = '#bbf7d0';
                    iconColor = '#34d399';
                    break;
                case NOTIFICATION_TYPES.error:
                    bgColor = '#991b1b';
                    textColor = '#fecaca';
                    iconColor = '#f87171';
                    break;
                case NOTIFICATION_TYPES.warning:
                    bgColor = '#92400e';
                    textColor = '#fef3c7';
                    iconColor = '#fbbf24';
                    break;
                case NOTIFICATION_TYPES.info:
                    bgColor = '#1e40af';
                    textColor = '#bfdbfe';
                    iconColor = '#60a5fa';
                    break;
            }
        }

        notification.style.backgroundColor = bgColor;
        notification.style.color = textColor;

        // Create icon
        const iconMap = {
            success: '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>',
            error: '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>',
            warning: '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>',
            info: '<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>'
        };

        const iconDiv = document.createElement('div');
        iconDiv.innerHTML = iconMap[type];
        iconDiv.style.marginRight = '0.75rem';
        iconDiv.style.flexShrink = '0';
        iconDiv.querySelector('svg') ? iconDiv.querySelector('svg').setAttribute('stroke', iconColor) : null;

        // Create message container
        const messageDiv = document.createElement('div');
        messageDiv.textContent = message;
        messageDiv.style.flexGrow = '1';

        // Create close button
        const closeBtn = document.createElement('button');
        closeBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>';
        closeBtn.style.background = 'none';
        closeBtn.style.border = 'none';
        closeBtn.style.color = 'inherit';
        closeBtn.style.width = '1.5rem';
        closeBtn.style.height = '1.5rem';
        closeBtn.style.display = 'flex';
        closeBtn.style.alignItems = 'center';
        closeBtn.style.justifyContent = 'center';
        closeBtn.style.cursor = 'pointer';
        closeBtn.style.opacity = '0.7';
        closeBtn.style.transition = 'opacity 0.2s';
        closeBtn.addEventListener('mouseenter', () => closeBtn.style.opacity = '1');
        closeBtn.addEventListener('mouseleave', () => closeBtn.style.opacity = '0.7');
        closeBtn.addEventListener('click', () => {
            this.remove(notification);
        });

        // Assemble notification
        notification.appendChild(iconDiv);
        notification.appendChild(messageDiv);
        notification.appendChild(closeBtn);

        // Add to container and track
        this.container.appendChild(notification);
        this.notifications.push({ element: notification, message, type });

        // Auto-remove after duration
        setTimeout(() => {
            this.remove(notification);
        }, duration);

        return notification;
    }

    /**
     * Remove a notification from the container and our tracking
     * @param {HTMLElement} element - The notification element to remove
     */
    remove(element) {
        if (element && element.parentNode) {
            element.style.animation = 'slideOut 0.3s ease-in forwards';
            // Actually remove after animation ends
            setTimeout(() => {
                if (element.parentNode) {
                    element.parentNode.removeChild(element);
                }
                // Remove from our tracking array
                const index = this.notifications.findIndex(n => n.element === element);
                if (index !== -1) {
                    this.notifications.splice(index, 1);
                }
            }, 300);
        }
    }

    /**
     * Show a success notification
     * @param {string} message
     * @param {number} duration
     */
    success(message, duration) {
        return this.show(message, NOTIFICATION_TYPES.success, duration);
    }

    /**
     * Show an error notification
     * @param {string} message
     * @param {number} duration
     */
    error(message, duration) {
        return this.show(message, NOTIFICATION_TYPES.error, duration);
    }

    /**
     * Show a warning notification
     * @param {string} message
     * @param {number} duration
     */
    warning(message, duration) {
        return this.show(message, NOTIFICATION_TYPES.warning, duration);
    }

    /**
     * Show an info notification
     * @param {string} message
     * @param {number} duration
     */
    info(message, duration) {
        return this.show(message, NOTIFICATION_TYPES.info, duration);
    }
}

// Create a global notification manager instance
const notificationManager = new NotificationManager();

// Export helper functions
window.showNotification = (message, type = 'info', duration = 5000) => {
    return notificationManager.show(message, type, duration);
};

window.showSuccessNotification = (message, duration) => {
    return notificationManager.success(message, duration);
};

window.showErrorNotification = (message, duration) => {
    return notificationManager.error(message, duration);
};

window.showWarningNotification = (message, duration) => {
    return notificationManager.warning(message, duration);
};

window.showInfoNotification = (message, duration) => {
    return notificationManager.info(message, duration);
};