function deletePost(postId) {
    if (!confirm('Are you sure you want to delete this post?')) {
        return;
    }

    fetch('/delete-post', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            post_id: parseInt(postId)
        }),
        credentials: 'include'
    })
    .then(response => {
        if (response.ok) {
            // Remove post from UI
            const postElement = document.getElementById(`post-${postId}`);
            if (postElement) {
                postElement.parentElement.remove();
            }
            // Redirect if on single post view
            if (window.location.search.includes('id=')) {
                window.location.href = '/';
            }
        } else {
            throw new Error('Failed to delete post');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        alert('Failed to delete post');
    });
}