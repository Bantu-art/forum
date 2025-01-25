// Theme toggle functionality
const themeToggle = document.getElementById('theme-toggle');
const html = document.documentElement;
const themeIcon = themeToggle.querySelector('i');

// saved theme preference
const savedTheme = localStorage.getItem('theme');
if (savedTheme) {
    html.setAttribute('data-theme', savedTheme);
    updateThemeIcon(savedTheme);
}

themeToggle.addEventListener('click', () => {
    const currentTheme = html.getAttribute('data-theme');
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    
    html.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeIcon(newTheme);
});

function updateThemeIcon(theme) {
    themeIcon.className = theme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
}

// Like, dislike, and comment functionality
document.querySelectorAll('.action-btn').forEach(button => {
    button.addEventListener('click', () => {
        const countSpan = button.querySelector('.count');
        let count = parseInt(countSpan.textContent);
        countSpan.textContent = count + 1;
        
        button.style.transform = 'scale(1.2)';
        setTimeout(() => {
            button.style.transform = 'scale(1)';
        }, 200);
    });
});

document.addEventListener('DOMContentLoaded', () => {
    document.body.style.transition = 'background-color 0.3s, color 0.3s';
});


// Image preview functionality
const imageInput = document.getElementById('post-image');
const imagePreview = document.getElementById('image-preview');

if (imageInput && imagePreview) {
    imageInput.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file && file.type.startsWith('image/')) {
            const reader = new FileReader();
            
            reader.onload = function(e) {
                imagePreview.innerHTML = `
                    <img src="${e.target.result}" alt="Preview" style="max-width: 100%; max-height: 300px; border-radius: 0.375rem;">
                `;
            };
            
            reader.readAsDataURL(file);
        }
    });
}

// Form submission handling

const createPostForm = document.getElementById('create-post-form');

if (createPostForm) {
    createPostForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const formData = new FormData(this);
        
        // For now, we'll just log it and redirect
        console.log('Title:', formData.get('title'));
        console.log('Description:', formData.get('description'));
        console.log('Image:', formData.get('image'));
        
        alert('Post created successfully!');
        
        // Redirect to home page
        window.location.href = 'index.html';
    });
}

// show password
document.getElementById('show-password').addEventListener('change', function() {
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirm_password');
    
    // Toggle type between "password" and "text"
    const type = this.checked ? 'text' : 'password';
    passwordInput.type = type;
    confirmPasswordInput.type = type;
});