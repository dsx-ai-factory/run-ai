
const themeName = "theme";
const themeAttributeName = "data-theme";

// Function to change the attribute of the <html> element
function changeDarkMode() {
    // Select the <html> element
    const htmlElement = document.documentElement;
    if (htmlElement.getAttribute(themeAttributeName) === 'light') {
        setTheme('dark');
    } else {
        setTheme('light');
    }
}

function setTheme(theme) {

    document.documentElement.setAttribute(themeAttributeName, theme);
    localStorage.setItem(themeName, theme);
}

const theme = localStorage.getItem(themeName) || 'dark';
setTheme(theme);

// Select the button element
const button = document.getElementById('changeDarkModeButton');

// Add event listener to the button to call changeHtmlAttribute on click
button.addEventListener('click', changeDarkMode);