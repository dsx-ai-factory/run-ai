import { policyModelUtil } from "./policy-model.util";
import type { IPolicyViewProperty } from "./policy-view.util";

describe("policy-model.util", () => {
  test("policyRulesToPreviewString will return normal rules in the right format", () => {
    const expected = "• required = false <br />• canEdit = true <br />";
    const rowRules = environmentRowsMock[0].rules;
    expect(policyModelUtil.policyRulesToPreviewString(rowRules)).toBe(expected);
  });

  test("policyRulesToPreviewString will return number rules in the right format", () => {
    const expected = "• canEdit = true <br />• min = 1 <br />• max = 5 <br />• step = 1 <br />";
    const rowRules = environmentRowsMock[3].rules;
    expect(policyModelUtil.policyRulesToPreviewString(rowRules)).toBe(expected);
  });

  test("policyDefaultToPreviewString will return array of strings in the correct format", () => {
    const expected = `• FOWNER <br />• KILL <br />`;
    const rowDefault = environmentRowsMock[1].default;
    expect(policyModelUtil.policyDefaultToPreviewString(rowDefault)).toBe(expected);
  });

  test("policyDefaultToPreviewString will return a string value in the correct format", () => {
    const expected = environmentRowsMock[0].default;
    const rowDefault = environmentRowsMock[0].default;
    expect(policyModelUtil.policyDefaultToPreviewString(rowDefault)).toBe(expected);
  });

  test("policyDefaultToPreviewString will an array of objects in the correct format", () => {
    const rowDefault = environmentRowsMock[2].default;
    const firstItemKeys = Object.keys(rowDefault[0]);
    const firstItemValues = Object.values(rowDefault[0]);
    const secondItemKeys = Object.keys(rowDefault[1]);
    const secondItemValues = Object.values(rowDefault[1]);
    const expected = `• ${firstItemKeys[0]} = ${firstItemValues[0]} <br />&nbsp; ${firstItemKeys[1]} = ${firstItemValues[1]} <br /><br />• ${secondItemKeys[0]} = ${secondItemValues[0]} <br />&nbsp; ${secondItemKeys[1]} = ${secondItemValues[1]} <br />`;
    expect(policyModelUtil.policyDefaultToPreviewString(rowDefault)).toBe(expected);
  });
});

const environmentRowsMock: IPolicyViewProperty[] = [
  {
    name: "image",
    sourceOfRule: {
      scope: "project",
      projectId: 1,
    },
    rules: {
      required: false,
      canEdit: true,
    },
    default: "ubuntu",
  },
  {
    name: "capabilities",
    sourceOfRule: {
      scope: "project",
      projectId: 1,
    },
    rules: {
      required: false,
      canEdit: true,
    },
    default: ["FOWNER", "KILL"],
  },
  {
    name: "environmentVariables",
    sourceOfRule: {
      scope: "project",
      projectId: 1,
    },
    rules: {
      canAdd: false,
    },
    default: [
      {
        name: "first-name",
        value: "first-val",
      },
      {
        name: "second-name",
        value: "second-val",
      },
    ],
  },
  {
    name: "runAsUid",
    sourceOfRule: {
      scope: "project",
      projectId: 1,
    },
    rules: {
      canEdit: true,
      min: 1,
      max: 5,
      step: 1,
    },
    default: 2,
  },
];
