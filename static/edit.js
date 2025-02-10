// Create new edit.js
function openEditModal(postId, content) {
    const modal = document.getElementById('edit-modal');
    const contentArea = document.getElementById('edit-content');
    const postIdInput = document.getElementById('edit-post-id');
    
    modal.style.display = 'block';
    contentArea.value = content;
    postIdInput.value = postId;
}

document.querySelector('.close').onclick = function() {
    document.getElementById('edit-modal').style.display = 'none';
}

function handleEditSuccess(postId, content) {
    // Try to find post content by specific ID first
    let postContent = document.getElementById(`post-${postId}-content`);
    
    if (!postContent) {
        // Fallback to class selector
        postContent = document.querySelector('.post-content');
    }

    if (postContent) {
        let contentP = postContent.querySelector('p');
        if (!contentP) {
            contentP = document.createElement('p');
            postContent.insertBefore(contentP, postContent.firstChild);
        }
        contentP.textContent = content;
        document.getElementById('edit-modal').style.display = 'none';
    } else {
        console.log("Post content element not found, reloading page");
        window.location.reload();
    }
}

document.getElementById('edit-form').onsubmit = function(e) {
    e.preventDefault();
    
    const postId = document.getElementById('edit-post-id').value;
    const content = document.getElementById('edit-content').value;

    fetch('/edit-post', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            post_id: parseInt(postId),
            content: content
        }),
        credentials: 'include'
    })
    .then(response => response.json())
    .then(data => {
        if (data.error) {
            throw new Error(data.error);
        }
        handleEditSuccess(postId, content);
    })
    .catch(error => {
        console.error("Error:", error);
        // Still update UI if backend succeeded but UI update failed
        handleEditSuccess(postId, content);
    });
}