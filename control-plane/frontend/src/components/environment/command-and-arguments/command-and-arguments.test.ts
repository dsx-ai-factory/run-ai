import { mount, type VueWrapper } from "@vue/test-utils";
import { Quasar } from "quasar";
import { describe } from "vitest";
import { CommandAndArguments } from ".";
import { errorMessages } from "@/common/error-message.constant";

describe("command-and-arguments.vue", () => {
  let commandAndArgumentsCmp: VueWrapper<InstanceType<typeof CommandAndArguments>>;

  beforeEach(() => {
    commandAndArgumentsCmp = mount(CommandAndArguments, {
      global: {
        plugins: [Quasar],
      },
      props: {},
    });
  });

  test("component is loading properly", () => {
    expect(commandAndArgumentsCmp.html()).toContain("Set a command and arguments");
  });

  test("should not show the command and args inputs by default", () => {
    expect(commandAndArgumentsCmp.find("[aid=command-input]").exists()).toBeFalsy();
    expect(commandAndArgumentsCmp.find("[aid=args-input]").exists()).toBeFalsy();
  });

  test("should show the command and args inputs when clicking the add command and args button", async () => {
    const addCommandAndArgsButton = commandAndArgumentsCmp.get('[aid="add-command-and-args-button"]');
    await addCommandAndArgsButton.trigger("click");
    expect(commandAndArgumentsCmp.find("[aid=command-input]").exists()).toBe(true);
    expect(commandAndArgumentsCmp.find("[aid=args-input]").exists()).toBe(true);
  });

  test("should not show the command and args inputs when clicking the close command and args button", async () => {
    const addCommandAndArgsButton = commandAndArgumentsCmp.get('[aid="add-command-and-args-button"]');
    await addCommandAndArgsButton.trigger("click");

    const closeCommandAndArgsButton = commandAndArgumentsCmp.get('[aid="close-command-and-args-button"]');
    await closeCommandAndArgsButton.trigger("click");

    expect(commandAndArgumentsCmp.find("[aid=command-input]").exists()).toBe(false);
    expect(commandAndArgumentsCmp.find("[aid=args-input]").exists()).toBe(false);
  });

  test("should not emit 'update' if nothing changed in the inputs", async () => {
    const addCommandAndArgsButton = commandAndArgumentsCmp.get('[aid="add-command-and-args-button"]');
    await addCommandAndArgsButton.trigger("click");
    expect(commandAndArgumentsCmp.emitted("update")).toBeFalsy();
  });

  test("should emit 'update' when filling the command input", async () => {
    const addCommandAndArgsButton = commandAndArgumentsCmp.get('[aid="add-command-and-args-button"]');
    await addCommandAndArgsButton.trigger("click");
    await commandAndArgumentsCmp.find("[aid=command-input]").setValue("command");
    expect(commandAndArgumentsCmp.emitted("update")).toBeTruthy();
  });

  test("should emit 'update' when filling the args input", async () => {
    const addCommandAndArgsButton = commandAndArgumentsCmp.get('[aid="add-command-and-args-button"]');
    await addCommandAndArgsButton.trigger("click");
    await commandAndArgumentsCmp.find("[aid=args-input]").setValue("--arg1");
    expect(commandAndArgumentsCmp.emitted("update")).toBeTruthy();
  });

  test("should fill the command input when getting the command in the props", async () => {
    await commandAndArgumentsCmp.setProps({ commandAndArgs: { command: "command", args: "" } });
    const commandInput = commandAndArgumentsCmp.find("[aid=command-input]");
    expect((commandInput.element as HTMLInputElement).value).toBe("command");
  });

  test("should fill the args input when getting the args in the props", async () => {
    await commandAndArgumentsCmp.setProps({ commandAndArgs: { command: "", args: "--arg1" } });
    const argsInput = commandAndArgumentsCmp.find("[aid=args-input]");
    expect((argsInput.element as HTMLInputElement).value).toBe("--arg1");
  });

  test("should clean the args and command values when clicking the close button", async () => {
    await commandAndArgumentsCmp.setProps({ commandAndArgs: { command: "command", args: "--arg1" } });
    const closeCommandAndArgsButton = commandAndArgumentsCmp.find('[aid="close-command-and-args-button"]');
    await closeCommandAndArgsButton.trigger("click");
    expect(commandAndArgumentsCmp.emitted("update")).toEqual([[{ command: "", args: "" }]]);
  });

  test("should return visible validation messages for invalid command and arguments", () => {
    expect(commandAndArgumentsCmp.vm.isValidCommand('python "unterminated')).toBe(errorMessages.ENTER_A_VALID_COMMAND);
    expect(commandAndArgumentsCmp.vm.isValidArgs('--name "unterminated')).toBe(errorMessages.ENTER_VALID_ARGUMENTS);
  });
});
