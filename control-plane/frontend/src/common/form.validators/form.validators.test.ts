import { describe, expect, it } from "vitest";
import {
  isEmpty,
  isEmptyObject,
  isValidDirectoryPath,
  validCommandAndArgsLine,
  haggingFaceTokenValid,
  huggingFaceModelNameValid,
  isValidDockerServerUrl,
  isValueLowerOrEqualToMax,
  isValidEmail,
  isValidPassword,
  isLowerCaseAndNumbersAndUnderscore,
  nimModelNameValidator,
  nimProfileValidator,
  isAlphaNumericOnly,
  isNotEmptyArray,
  isValidPrivateUrl,
} from "./form.validators";

describe("form.validators", () => {
  it("should validate directory", () => {
    expect(isValidDirectoryPath("/")).toBeTruthy();
    expect(isValidDirectoryPath("/test/ss")).toBeTruthy();
    expect(isValidDirectoryPath("test")).toBeFalsy();
    expect(isValidDirectoryPath("./test")).toBeFalsy();
    expect(isValidDirectoryPath("../test")).toBeFalsy();
    expect(isValidDirectoryPath("/~")).toBeFalsy();
  });

  it("should validate command and arguments", () => {
    expect(validCommandAndArgsLine('args1"args2')).toBe(false);
    expect(validCommandAndArgsLine("args1-\"'args2")).toBe(false);
    expect(validCommandAndArgsLine('args1"args2"')).toBe(true);
    expect(validCommandAndArgsLine('args1 "args2"')).toBe(true);
    expect(validCommandAndArgsLine("args1/args2")).toBe(true);
    expect(validCommandAndArgsLine("args1;args2")).toBe(true);
    expect(validCommandAndArgsLine("")).toBe(false);
    expect(validCommandAndArgsLine(null)).toBe(true);
  });

  it("should check if object is empty", () => {
    expect(isEmptyObject({})).toBeTruthy();
    expect(isEmptyObject({ key: 1, value: 1 })).toBeFalsy();
    expect(isEmptyObject(null)).toBeTruthy();
    expect(isEmptyObject(undefined)).toBeTruthy();
  });

  it("should check if string is empty", () => {
    expect(isEmpty("")).toBeTruthy();
    expect(isEmpty("string")).toBeFalsy();
    expect(isEmpty("s")).toBeFalsy();
    expect(isEmpty(null)).toBeTruthy();
    expect(isEmpty(undefined)).toBeTruthy();
  });

  describe("haggingFaceTokenValid", () => {
    it("should match valid tokens", () => {
      const validTokens = ["hf_someValidToken"];

      validTokens.forEach((token) => {
        expect(haggingFaceTokenValid(token)).toBe(true);
      });
    });

    it("should not match invalid tokens", () => {
      expect(haggingFaceTokenValid("h_notAValidToken")).toBe(false);
      expect(haggingFaceTokenValid("f_notAValidToken")).toBe(false);
      expect(haggingFaceTokenValid("_notAValidToken")).toBe(false);
      expect(haggingFaceTokenValid("hf_someValidToken")).toBe(true);
      expect(haggingFaceTokenValid("hf_notAValidToken_")).toBe(false);
    });
  });

  describe("huggingFaceModelNameValid", () => {
    it("should return true for valid model names", () => {
      expect(huggingFaceModelNameValid("bert-base-uncased")).toBe(true);
      expect(huggingFaceModelNameValid("gpt2")).toBe(true);
      expect(huggingFaceModelNameValid("gpt2/")).toBe(false);
      expect(huggingFaceModelNameValid("gpt2/f")).toBe(true);
      expect(huggingFaceModelNameValid("transformers/distilbert-base-uncased")).toBe(true);
      expect(huggingFaceModelNameValid("transformers/distilbert-base-cased")).toBe(true);
      expect(huggingFaceModelNameValid("abc123/def.456")).toBe(true);
    });

    it("should return false for invalid model names", () => {
      expect(huggingFaceModelNameValid("invalid_model!")).toBe(false);
      expect(huggingFaceModelNameValid(" ")).toBe(false);
      expect(huggingFaceModelNameValid("model/name/with spaces")).toBe(false);
      expect(huggingFaceModelNameValid("model/name/with_underscore_and_invalid-character$")).toBe(false);
    });
  });

  describe("isValidDockerServerUrl", () => {
    it("should return true for valid docker server urls", () => {
      const validUrls = [
        "http://example.com",
        "https://example.com",
        "https://example.com/path",
        "https://example.com/path/",
        "https://example.com/path/v1",
        "http://example.com:8080",
        "https://example.com:8080",
        "example.com",
        "example.com:8080",
      ];
      validUrls.forEach((url) => {
        expect(isValidDockerServerUrl(url)).toBe(true);
      });
    });

    it("should return false for invalid docker server urls", () => {
      const invalidUrls = ["ftp://example.com", "http://example", "just-a-string", "43", "http", "https"];
      invalidUrls.forEach((url) => {
        expect(isValidDockerServerUrl(url)).toBe(false);
      });
    });
  });

  describe("isValueLowerOrEqualToMax", () => {
    it("should return true when value is lower than the max", () => {
      expect(isValueLowerOrEqualToMax(1, 2)).toBe(true);
    });

    it("should return true when value is equal to max", () => {
      expect(isValueLowerOrEqualToMax(2, 2)).toBe(true);
    });

    it("should return false when value is higher than the max", () => {
      expect(isValueLowerOrEqualToMax(3, 2)).toBe(false);
    });
  });
});

describe("isValidEmail", () => {
  // Valid email cases
  it("should return true for a simple valid email", () => {
    expect(isValidEmail("example@example.com")).toBe(true);
  });

  it("should return true for an email with subdomains", () => {
    expect(isValidEmail("user.name@sub.example.co.uk")).toBe(true);
  });

  it("should return true for an email with allowed special characters", () => {
    expect(isValidEmail("user_name+alias@example.com")).toBe(true);
  });

  it('should return false for an email without "@" symbol', () => {
    expect(isValidEmail("example.com")).toBe(false);
  });

  it("should return false for an email with consecutive dots in local part", () => {
    expect(isValidEmail("user..name@example.com")).toBe(false);
  });

  it("should return false for an email with consecutive dots in domain part", () => {
    expect(isValidEmail("user@sub..example.com")).toBe(false);
  });

  it("should return false for an email with a dot at the beginning of the local part", () => {
    expect(isValidEmail(".username@example.com")).toBe(false);
  });

  it("should return false for an email with a dot at the end of the local part", () => {
    expect(isValidEmail("username.@example.com")).toBe(false);
  });

  it("should return false for an email with a space", () => {
    expect(isValidEmail("user name@example.com")).toBe(false);
  });

  it("should return false for an email without a top-level domain", () => {
    expect(isValidEmail("username@example")).toBe(false);
  });

  it("should return false for an email with invalid characters", () => {
    expect(isValidEmail("username@exa!mple.com")).toBe(false);
  });

  it("should return false for an email with an invalid top-level domain", () => {
    expect(isValidEmail("username@example.c")).toBe(false);
  });

  it("should return false for an empty string", () => {
    expect(isValidEmail("")).toBe(false);
  });

  it("should return false for null", () => {
    expect(isValidEmail(null as unknown as string)).toBe(false);
  });

  it("should return false for undefined", () => {
    expect(isValidEmail(undefined as unknown as string)).toBe(false);
  });
});

describe("isNotEmptyArray", () => {
  it("should return true for a non-empty number array", () => {
    expect(isNotEmptyArray([1, 2, 3])).toBe(true);
  });

  it("should return true for a non-empty string array", () => {
    expect(isNotEmptyArray(["1", "2", "3"])).toBe(true);
  });

  it("should return false for an empty array", () => {
    expect(isNotEmptyArray([])).toBe(false);
  });
});

describe("Password Validator", () => {
  test("valid password", () => {
    expect(isValidPassword("Valid123!")).toBe(true);
  });

  test("missing lowercase letter", () => {
    expect(isValidPassword("INVALID123!")).toBe(false);
  });

  test("missing uppercase letter", () => {
    expect(isValidPassword("invalid123!")).toBe(false);
  });

  test("missing digit", () => {
    expect(isValidPassword("InvalidPass!")).toBe(false);
  });

  test("missing special character", () => {
    expect(isValidPassword("Invalid123")).toBe(false);
  });

  test("too short password", () => {
    expect(isValidPassword("Inv1!")).toBe(false);
  });

  test("only special characters", () => {
    expect(isValidPassword("!@#$%^&*")).toBe(false);
  });

  test("valid password with different valid combinations", () => {
    expect(isValidPassword("Password1@")).toBe(true);
    expect(isValidPassword("AnotherPass2#")).toBe(true);
    expect(isValidPassword("TestPassword3$")).toBe(true);
  });
});

describe("isValidRunaiAppNameFormat", () => {
  // Valid email cases
  it("should return true for a simple valid email", () => {
    expect(isLowerCaseAndNumbersAndUnderscore("app_-lication")).toBe(true);
    expect(isLowerCaseAndNumbersAndUnderscore("application")).toBe(true);
    expect(isLowerCaseAndNumbersAndUnderscore("-Application")).toBe(false);
    expect(isLowerCaseAndNumbersAndUnderscore("_App_-lication")).toBe(false);
    expect(isLowerCaseAndNumbersAndUnderscore("Application")).toBe(false);
    expect(isLowerCaseAndNumbersAndUnderscore("_App@-lication")).toBe(false);
    expect(isLowerCaseAndNumbersAndUnderscore("1")).toBe(false);
  });
});

describe("nimModelNameValidator", () => {
  it("should return true for valid model names", () => {
    const validNames = ["model123", "model_123", "model-123", "model.123", "Model-Name/Path", "path/to.model"];

    validNames.forEach((name) => {
      expect(nimModelNameValidator(name)).toBe(true);
    });
  });

  it("should return false for invalid model names", () => {
    const invalidNames = [
      "model name", // contains spaces
      "model@name", // contains special character @
      "model#name", // contains special character #
      "model$name", // contains special character $
      "model!name", // contains special character !
      "model%name", // contains special character %
      "model^name", // contains special character ^
      "", // empty string
      " model123", // leading space
      "model123 ", // trailing space
    ];

    invalidNames.forEach((name) => {
      expect(nimModelNameValidator(name)).toBe(false);
    });
  });

  it("should handle edge cases", () => {
    expect(nimModelNameValidator("")).toBe(false); // empty string
    expect(nimModelNameValidator(".")).toBe(true); // single dot
    expect(nimModelNameValidator("a")).toBe(true); // single character
    expect(nimModelNameValidator("-")).toBe(true); // single hyphen
    expect(nimModelNameValidator("/")).toBe(true); // single slash
    expect(nimModelNameValidator("_")).toBe(true); // single underscore
    expect(nimModelNameValidator("A")).toBe(true); // single uppercase letter
    expect(nimModelNameValidator("1")).toBe(true); // single digit
  });
});

describe("nimProfileValidator", () => {
  it("should return true for valid profile names", () => {
    const validNames = [
      "profile123",
      "profile_123",
      "profile-123",
      "ProfileName",
      "PROFILE_NAME",
      "Profile-Name123",
      "123Profile",
    ];

    validNames.forEach((name) => {
      expect(nimProfileValidator(name)).toBe(true);
    });
  });

  it("should return false for invalid profile names", () => {
    const invalidNames = [
      "profile name", // contains spaces
      "profile@name", // contains special character @
      "profile#name", // contains special character #
      "profile$name", // contains special character $
      "profile!name", // contains special character !
      "profile%name", // contains special character %
      "profile^name", // contains special character ^
      "profile.name", // contains special character .
      "profile/name", // contains special character /
      "", // empty string
      " profile123", // leading space
      "profile123 ", // trailing space
    ];

    invalidNames.forEach((name) => {
      expect(nimProfileValidator(name)).toBe(false);
    });
  });

  it("should handle edge cases", () => {
    expect(nimProfileValidator("")).toBe(false); // empty string
    expect(nimProfileValidator("a")).toBe(true); // single lowercase character
    expect(nimProfileValidator("A")).toBe(true); // single uppercase character
    expect(nimProfileValidator("1")).toBe(true); // single digit
    expect(nimProfileValidator("_")).toBe(true); // single underscore
    expect(nimProfileValidator("-")).toBe(true); // single hyphen
    expect(nimProfileValidator("123")).toBe(true); // numeric string
    expect(nimProfileValidator("A_B-C")).toBe(true); // mixed valid characters
  });
});

describe("isAlphaNumericOnly", () => {
  it("should return true for strings containing only letters and numbers", () => {
    expect(isAlphaNumericOnly("abc")).toBe(true);
    expect(isAlphaNumericOnly("ABC")).toBe(true);
    expect(isAlphaNumericOnly("123")).toBe(true);
    expect(isAlphaNumericOnly("aBc123")).toBe(true);
  });

  it("should return false for strings containing special characters", () => {
    expect(isAlphaNumericOnly("abc!")).toBe(false);
    expect(isAlphaNumericOnly("123@")).toBe(false);
    expect(isAlphaNumericOnly("aBc#123")).toBe(false);
    expect(isAlphaNumericOnly("aBc_")).toBe(false);
    expect(isAlphaNumericOnly("aBc.")).toBe(false);
    expect(isAlphaNumericOnly("aBc/")).toBe(false);
    expect(isAlphaNumericOnly("aBc$")).toBe(false);
    expect(isAlphaNumericOnly("aBc&")).toBe(false);
    expect(isAlphaNumericOnly("hello world")).toBe(false); // Space is a special character
  });

  it("should return false for strings containing spaces", () => {
    expect(isAlphaNumericOnly(" ")).toBe(false);
    expect(isAlphaNumericOnly("a b c")).toBe(false);
  });

  it("should return false for empty strings", () => {
    expect(isAlphaNumericOnly("")).toBe(false);
  });

  it("should handle numeric-only strings", () => {
    expect(isAlphaNumericOnly("123456")).toBe(true);
  });

  it("should handle alphabetic-only strings", () => {
    expect(isAlphaNumericOnly("abcDEF")).toBe(true);
  });
});

describe("isValidPrivateUrl", () => {
  describe("Git SSH URLs", () => {
    it("should validate Git SSH URLs", () => {
      expect(isValidPrivateUrl("git@github.com:username/repo.git")).toBe(true);
      expect(isValidPrivateUrl("git@gitlab.com:group/project.git")).toBe(true);
      expect(isValidPrivateUrl("git@bitbucket.org:username/repo.git")).toBe(true);
    });

    it("should validate Git SSH URLs without .git extension", () => {
      expect(isValidPrivateUrl("git@github.com:username/repo")).toBe(true);
      expect(isValidPrivateUrl("git@gitlab.com:group/project")).toBe(true);
    });

    it("should validate Git SSH URLs with special characters in username", () => {
      expect(isValidPrivateUrl("user-name@github.com:username/repo.git")).toBe(true);
      expect(isValidPrivateUrl("user_name@github.com:username/repo.git")).toBe(true);
    });
  });

  describe("HTTPS URLs", () => {
    it("should validate standard HTTPS URLs", () => {
      expect(isValidPrivateUrl("https://github.com/username/repo.git")).toBe(true);
      expect(isValidPrivateUrl("https://gitlab.com/group/project.git")).toBe(true);
      expect(isValidPrivateUrl("https://bitbucket.org/username/repo.git")).toBe(true);
    });

    it("should validate URLs with query parameters", () => {
      expect(isValidPrivateUrl("https://github.com/username/repo?ref=main")).toBe(true);
      expect(isValidPrivateUrl("https://gitlab.com/group/project?token=abc123")).toBe(true);
    });

    it("should validate URLs with fragments", () => {
      expect(isValidPrivateUrl("https://github.com/username/repo#readme")).toBe(true);
      expect(isValidPrivateUrl("https://gitlab.com/group/project#section")).toBe(true);
    });

    it("should validate URLs with special characters in paths", () => {
      expect(isValidPrivateUrl("https://github.com/username/repo-name.git")).toBe(true);
      expect(isValidPrivateUrl("https://gitlab.com/group/project_name.git")).toBe(true);
      expect(isValidPrivateUrl("https://github.com/username/repo.with.dots.git")).toBe(true);
    });
  });

  describe("HTTP URLs", () => {
    it("should validate HTTP URLs", () => {
      expect(isValidPrivateUrl("http://github.com/username/repo.git")).toBe(true);
      expect(isValidPrivateUrl("http://gitlab.com/group/project.git")).toBe(true);
    });
  });

  describe("URLs without protocol", () => {
    it("should validate URLs without protocol", () => {
      expect(isValidPrivateUrl("github.com/username/repo.git")).toBe(true);
      expect(isValidPrivateUrl("gitlab.com/group/project.git")).toBe(true);
    });
  });

  describe("URLs with ports", () => {
    it("should validate URLs with ports", () => {
      expect(isValidPrivateUrl("https://github.com:8080/username/repo.git")).toBe(true);
      expect(isValidPrivateUrl("git@github.com:8080:username/repo.git")).toBe(true);
    });
  });

  describe("Invalid URLs", () => {
    it("should reject invalid URLs", () => {
      expect(isValidPrivateUrl("not-a-url")).toBe(false);
      expect(isValidPrivateUrl("http://")).toBe(false);
      expect(isValidPrivateUrl("https://")).toBe(false);
      expect(isValidPrivateUrl("git@")).toBe(false);
      expect(isValidPrivateUrl("@github.com")).toBe(false);
    });

    it("should reject malformed Git SSH URLs", () => {
      expect(isValidPrivateUrl("git@github.com")).toBe(false);
      expect(isValidPrivateUrl("git@github.com:")).toBe(false);
      expect(isValidPrivateUrl("git@github.com:username/")).toBe(false);
    });
  });
});
