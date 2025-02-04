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

document.addEventListener('DOMContentLoaded', function() {
    document.querySelectorAll('.action-btn').forEach(button => {
        button.addEventListener('click', function() {
            const postId = this.getAttribute('data-post-id');
            const action = this.getAttribute('data-action');
            handleReaction(postId, action);
        });
    });
});

function handleReaction(postId, action) {
    const sessionToken = document.cookie
        .split('; ')
        .find(row => row.startsWith('session_token='))
        ?.split('=')[1];

    console.log("Session token found:", sessionToken);

    if (!sessionToken) {
        window.location.href = '/signin';
        return;
    }

    fetch('/react', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            post_id: parseInt(postId),
            reaction_type: action
        }),
        credentials: 'include'
    })
    .then(res => res.json())
    .then(data => {
        document.querySelector(`#likes-${postId}`).textContent = data.likes;
        document.querySelector(`#dislikes-${postId}`).textContent = data.dislikes;
    })
    .catch(error => console.error('Error:', error));
}


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
        

        console.log('Title:', formData.get('title'));
        console.log('Description:', formData.get('description'));
        console.log('Image:', formData.get('image'));
        
        alert('Post created successfully!');
        

        window.location.href = 'index.html';
    });
}


document.getElementById('show-password').addEventListener('change', function() {
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirm_password');
    

    const type = this.checked ? 'text' : 'password';
    passwordInput.type = type;
    confirmPasswordInput.type = type;
});