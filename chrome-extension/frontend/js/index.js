// VMS Homepage - Event Handlers

document.addEventListener('DOMContentLoaded', function() {
    // Navigation links
    const navLinks = document.querySelectorAll('.nav-links a[href^="#"]');
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            const href = this.getAttribute('href');
            if (href.startsWith('#')) {
                const target = document.querySelector(href);
                if (target) {
                    e.preventDefault();
                    target.scrollIntoView({ behavior: 'smooth' });
                }
            }
        });
    });

    // CTA Buttons
    const ctaButtons = document.querySelectorAll('[data-action="login"]');
    ctaButtons.forEach(button => {
        button.addEventListener('click', function() {
            window.location.href = 'pages/login.html';
        });
    });

    // Scroll to features button
    const scrollButton = document.querySelector('[data-action="scroll-features"]');
    if (scrollButton) {
        scrollButton.addEventListener('click', function() {
            document.getElementById('features').scrollIntoView({ behavior: 'smooth' });
        });
    }

    // Smooth scroll for all anchor links
    document.querySelectorAll('a[href*="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            const href = this.getAttribute('href');
            if (href.startsWith('#') && href.length > 1) {
                e.preventDefault();
                const target = document.querySelector(href);
                if (target) {
                    target.scrollIntoView({ behavior: 'smooth' });
                }
            }
        });
    });
});

// Utility function to navigate
function navigateTo(path) {
    window.location.href = path;
}
