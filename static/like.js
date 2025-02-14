function handleReaction(event) {
    event.preventDefault();
    event.stopPropagation();

    const button = event.currentTarget;
    const postID = button.getAttribute("data-post-id");
    const action = button.getAttribute("data-action");
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
        // First check if the response is JSON
        const contentType = response.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
            // Not JSON, probably redirected to signin page
            window.location.href = '/signin';
            throw new Error('Session expired. Please sign in again.');
        }
        
        if (response.status === 401) {
            window.location.href = '/signin';
            throw new Error('Please sign in to react to posts');
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
        if (!error.message.includes("Please sign in")) {
            alert(error.message);
        }
    });
}
// Handler for comment reactions
function handleCommentReaction(event) {
    event.preventDefault();
    event.stopPropagation();

    const button = event.currentTarget;
    const commentID = button.getAttribute("data-comment-id");
    const action = button.getAttribute("data-action");
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
        credentials: 'include'
    })
    .then(response => {
        const contentType = response.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
            window.location.href = '/signin';
            throw new Error('Session expired. Please sign in again.');
        }

        if (response.status === 401) {
            window.location.href = '/signin';
            throw new Error('Please sign in to react to comments');
        }
        return response.json();
    })
    .then(data => {
        if (data.error) {
            throw new Error(data.error);
        }
        const likesElement = document.getElementById(`comment-likes-${commentID}`);
        const dislikesElement = document.getElementById(`comment-dislikes-${commentID}`);
        
        if (likesElement && dislikesElement) {
            likesElement.textContent = data.likes;
            dislikesElement.textContent = data.dislikes;
        }
    })
    .catch(error => {
        console.error("Error:", error);
        if (!error.message.includes("Please sign in")) {
            alert(error.message);
        }
    });
}
// Comment editing functionality
function editComment(commentId) {
    const contentDiv = document.getElementById(`comment-content-${commentId}`);
    if (!contentDiv) return;
    
    // Check if we're already editing
    if (contentDiv.querySelector('.edit-comment-form')) {
        return;
    }
    
    const currentContent = contentDiv.textContent.trim();
    
    // Create edit form
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = '/editcomment';
    form.className = 'edit-comment-form';
    
    // Create textarea
    const textarea = document.createElement('textarea');
    textarea.name = 'content';
    textarea.value = currentContent;
    textarea.required = true;
    textarea.className = 'comment-input';
    
    // Create hidden input for comment ID
    const hiddenInput = document.createElement('input');
    hiddenInput.type = 'hidden';
    hiddenInput.name = 'comment_id';
    hiddenInput.value = commentId;
    
    // Create buttons container
    const buttonsDiv = document.createElement('div');
    buttonsDiv.className = 'edit-buttons';
    
    // Create save button
    const saveButton = document.createElement('button');
    saveButton.type = 'submit';
    saveButton.className = 'save-btn';
    saveButton.textContent = 'Save';
    
    // Create cancel button
    const cancelButton = document.createElement('button');
    cancelButton.type = 'button';
    cancelButton.className = 'cancel-btn';
    cancelButton.textContent = 'Cancel';
    cancelButton.onclick = (e) => {
        e.preventDefault();
        contentDiv.innerHTML = currentContent;
    };
    
    // Assemble the form
    buttonsDiv.appendChild(saveButton);
    buttonsDiv.appendChild(cancelButton);
    form.appendChild(textarea);
    form.appendChild(hiddenInput);
    form.appendChild(buttonsDiv);
    
    // Replace content with form
    contentDiv.innerHTML = '';
    contentDiv.appendChild(form);
    
    // Focus the textarea
    textarea.focus();
    textarea.setSelectionRange(textarea.value.length, textarea.value.length);
}

// Comment deletion confirmation
function confirmDeleteComment(event, commentId) {
    event.preventDefault();
    
    if (confirm('Are you sure you want to delete this comment? This action cannot be undone.')) {
        const form = event.target.closest('form');
        if (form) {
            form.submit();
        }
    }
}

// Attach event listeners for post reaction buttons
document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", handleReaction);
});

// Attach event listeners for comment reaction buttons
document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
    button.addEventListener("click", handleCommentReaction);
});

document.addEventListener('DOMContentLoaded', function() {
    const hamburgerBtn = document.querySelector('.hamburger-btn');
    const navRight = document.querySelector('.nav-right');
    const overlay = document.querySelector('.mobile-menu-overlay');
    const menuToggles = document.querySelectorAll('.menu-toggle-btn');

    // Toggle mobile menu
    hamburgerBtn.addEventListener('click', () => {
        navRight.classList.toggle('active');
        overlay.classList.toggle('active');
        document.body.style.overflow = navRight.classList.contains('active') ? 'hidden' : '';
    });

    // Close menu when clicking overlay
    overlay.addEventListener('click', () => {
        navRight.classList.remove('active');
        overlay.classList.remove('active');
        document.body.style.overflow = '';
    });

    // Toggle categories and filters sections
    menuToggles.forEach(toggle => {
        toggle.addEventListener('click', () => {
            const content = toggle.nextElementSibling;
            toggle.classList.toggle('active');
            content.classList.toggle('active');
        });
    });
});
