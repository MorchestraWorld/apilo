// Apilo Documentation Interactive Features

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', function() {
    initializeSearch();
    initializeCodeHighlighting();
    initializeSmoothScrolling();
    updateBuildDate();
    initializeTheme();
});

// Search functionality
function initializeSearch() {
    const searchInput = document.getElementById('doc-search');
    if (!searchInput) return;

    searchInput.addEventListener('input', function(e) {
        const query = e.target.value.toLowerCase();
        filterNavigation(query);
    });
}

function filterNavigation(query) {
    const navSections = document.querySelectorAll('.nav-section');

    navSections.forEach(section => {
        const links = section.querySelectorAll('a');
        let hasVisibleLinks = false;

        links.forEach(link => {
            const text = link.textContent.toLowerCase();
            if (text.includes(query)) {
                link.style.display = 'block';
                hasVisibleLinks = true;
            } else {
                link.style.display = query ? 'none' : 'block';
            }
        });

        // Hide section header if no visible links
        const header = section.querySelector('h3');
        if (header) {
            header.style.display = hasVisibleLinks || !query ? 'block' : 'none';
        }
    });
}

// Code syntax highlighting
function initializeCodeHighlighting() {
    if (typeof hljs !== 'undefined') {
        hljs.highlightAll();
    }
}

// Smooth scrolling for anchor links
function initializeSmoothScrolling() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });

                // Update active nav link
                updateActiveNavLink(this);
            }
        });
    });
}

function updateActiveNavLink(clickedLink) {
    // Remove active class from all links
    document.querySelectorAll('.nav-section a').forEach(link => {
        link.classList.remove('active');
    });

    // Add active class to clicked link
    clickedLink.classList.add('active');
}

// Update build date
function updateBuildDate() {
    const buildDateElement = document.getElementById('build-date');
    if (buildDateElement) {
        const now = new Date();
        buildDateElement.textContent = now.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
        });
    }
}

// Theme management (dark/light)
function initializeTheme() {
    // Check for saved theme preference or default to dark
    const currentTheme = localStorage.getItem('theme') || 'dark';
    document.documentElement.setAttribute('data-theme', currentTheme);
}

function toggleTheme() {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
}

// Scroll spy for table of contents
window.addEventListener('scroll', function() {
    const sections = document.querySelectorAll('.section[id]');
    const navLinks = document.querySelectorAll('.nav-section a');

    let currentSection = '';

    sections.forEach(section => {
        const sectionTop = section.offsetTop;
        const sectionHeight = section.clientHeight;

        if (window.pageYOffset >= sectionTop - 100) {
            currentSection = section.getAttribute('id');
        }
    });

    navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === '#' + currentSection) {
            link.classList.add('active');
        }
    });
});

// Copy code to clipboard
function initializeCodeCopy() {
    const codeBlocks = document.querySelectorAll('.code-block');

    codeBlocks.forEach(block => {
        const copyButton = document.createElement('button');
        copyButton.className = 'copy-button';
        copyButton.textContent = 'Copy';

        copyButton.addEventListener('click', function() {
            const code = block.querySelector('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                copyButton.textContent = 'Copied!';
                setTimeout(() => {
                    copyButton.textContent = 'Copy';
                }, 2000);
            });
        });

        block.appendChild(copyButton);
    });
}

// Initialize copy buttons when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeCodeCopy);
} else {
    initializeCodeCopy();
}
