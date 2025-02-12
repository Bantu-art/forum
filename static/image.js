document.getElementById('image-input').addEventListener('change', function(e) {
    const file = e.target.files[0];
    const preview = document.getElementById('image-preview');
    
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            preview.innerHTML = `<img src="${e.target.result}" style="max-width: 100%; height: auto;">`;
        }
        reader.readAsDataURL(file);
    }
});


function validateCategories() {
    const checkboxes = document.querySelectorAll('input[name="categories[]"]');
    let checkedOne = Array.prototype.slice.call(checkboxes).some(x => x.checked);
    if (!checkedOne) {
        document.getElementById('category-error').style.display = 'block';
        return false;
    }
    return true;
}
