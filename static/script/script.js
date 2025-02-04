
const themeToggle = document.getElementById('theme-toggle');
const html = document.documentElement;
const themeIcon = themeToggle.querySelector('i');


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


imageInput.addEventListener('change', function () {
    const file = this.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function (e) {
            imagePreview.innerHTML = `<img src="${e.target.result}" alt="Preview" />`;
        };
        reader.readAsDataURL(file);
    } else {
        imagePreview.innerHTML = `<i class="fas fa-image"></i><p>Click to upload image</p>`;
    }
});


const form = document.getElementById('create-post-form');
form.addEventListener('submit', function (e) {
    e.preventDefault();
    const title = document.getElementById('post-title').value;
    const description = document.getElementById('post-description').value;
    const categories = Array.from(document.getElementById('post-categories').selectedOptions)
        .map(option => option.value);
    const image = document.getElementById('post-image').files[0];

    console.log('Title:', title);
    console.log('Description:', description);
    console.log('Categories:', categories);
    console.log('Image:', image);


    alert('Post created successfully!');
    window.location.href = '/';
});
async function loadPosts(category) {
const response = await fetch(`/posts?category=${category}`);
const posts = await response.json();
const postsContainer = document.getElementById('posts-container');
    postsContainer.innerHTML = '';

    posts.forEach(post => {
                const postCard = `
                    <div class="post-card">
                        <a href="/?id=${post.ID}" class="post-link"></a>
                        <div class="post-header">
                            <div class="avatar"></div>
                            <div class="post-info">
                                <h3>${post.Title}</h3>
                                <span class="timestamp">${post.PostTime}</span>
                            </div>
                        </div>
                        <div class="post-content">
                            <p>${post.Content}</p>
                            ${post.ImagePath ? `<img src="${post.ImagePath}" alt="Post image" class="post-image">` : ''}
                        </div>
                        <div class="post-footer">
                            <button class="action-btn like-btn" onclick="event.stopPropagation();">
                                <i class="fas fa-heart"></i>
                                <span class="count">${post.Likes}</span>
                            </button>
                            <button class="action-btn dislike-btn" onclick="event.stopPropagation();">
                                <i class="fas fa-thumbs-down"></i>
                                <span class="count">${post.Dislikes}</span>
                            </button>
                            <button class="action-btn comment-btn" onclick="event.stopPropagation();">
                                <i class="fas fa-comment"></i>
                                <span class="count">${post.Comments}</span>
                            </button>
                        </div>
                    </div>
                `;
                postsContainer.innerHTML += postCard;
            });
   
     }
     function handleReaction(event) {
        event.preventDefault();
    
        event.stopPropagation(); // Prevent post link click when clicking like/dislike
    
        const button = event.currentTarget;
        const postID = button.getAttribute("data-post-id");
        const action = button.getAttribute("data-action");
    
        // Check if user is logged in by looking for session cookie
        const hasSession = document.cookie.includes('session_token=');
        if (!hasSession) {
            window.location.href = '/signin';
            return;
        }
    
        const like = action === "like" ? 1 : 0;
    
        fetch("/react", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                post_id: parseInt(postID),
                like: like,
            }),
            credentials: 'include'
        })
        .then(response => {
            if (response.status === 401) {
    
                window.location.href = '/signin';
                throw new Error('Please log in to react to posts');
            }
            return response.json();
        })
        .then(data => {
            if (data.error) {
                throw new Error(data.error);
            }
            const likesElement = document.getElementById(`likes-${postID}`);
            const dislikesElement = document.getElementById(`dislikes-${postID}`);
            
            if (likesElement && dislikesElement) {
                likesElement.textContent = data.likes;
                dislikesElement.textContent = data.dislikes;
            }
        })
        .catch(error => {
            console.error("Error:", error);
            alert(error.message);
        });
    }
    
    // Handler for comment reactions (modified to mirror the post reaction handler)
    function handleCommentReaction(event) {
        event.stopPropagation(); // Prevent any unwanted propagation
    
        const button = event.currentTarget;
        const commentID = button.getAttribute("data-comment-id");
        const action = button.getAttribute("data-action");
    
        // Check if user is logged in by looking for session cookie
        const hasSession = document.cookie.includes('session_token=');
        if (!hasSession) {
            window.location.href = '/signin';
            return;
        }
    
        const like = action === "like" ? 1 : 0;
    
        fetch("/commentreact", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                comment_id: parseInt(commentID),
                like: like,
            }),
            credentials: 'include' // Ensure cookies are sent
        })
        .then(response => {
            if (response.status === 401) {
                window.location.href = '/signin';
                throw new Error('Please log in to react to comments');
            }
            return response.json();
        })
        .then(data => {
            if (data.error) {
                throw new Error(data.error);
            }
            // Update the comment's like and dislike counts on the page
            const likesElement = document.getElementById(`comment-likes-${commentID}`);
            const dislikesElement = document.getElementById(`comment-dislikes-${commentID}`);
            console.log(likesElement)
            
            if (likesElement && dislikesElement) {
                likesElement.textContent = data.likes;
                dislikesElement.textContent = data.dislikes;
            }
        })
        .catch(error => {
            console.error("Error:", error);
            alert(error.message);
        });
    }
    
    // Attach event listeners for post reaction buttons
    document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
        button.addEventListener("click", handleReaction);
    });
    
    // Attach event listeners for comment reaction buttons
    document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
        button.addEventListener("click", handleCommentReaction);
    });
    