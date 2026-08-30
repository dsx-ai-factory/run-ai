export const formUtil = {
  enterKeySubmissionHandler,
};
/**
 * Handles Enter key submission for forms.
 * @param {KeyboardEvent} event - The keyboard event to handle.
 */
async function enterKeySubmissionHandler(event: KeyboardEvent, onSubmit: () => void | Promise<void>) {
  // Only handle Enter key
  if (event.key !== "Enter") return;

  // Don't handle if user is typing in an input field
  const target = event.target as HTMLElement;
  if (
    target.tagName === "INPUT" ||
    target.tagName === "TEXTAREA" ||
    target.tagName === "SELECT" ||
    target.isContentEditable ||
    target.closest("input") ||
    target.closest("textarea") ||
    target.closest(".q-field") ||
    target.closest(".q-select")
  ) {
    return;
  }

  // Prevent default form submission behavior
  event.preventDefault();
  event.stopPropagation();

  // Trigger submit callback
  await onSubmit();
}
