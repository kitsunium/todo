import { describe, expect, it } from "vitest";
import { ApiError, NETWORK_ERROR, toApiError } from "../src/api/client";
import { errorMessage, fieldErrors, isUnauthenticated } from "../src/api/errors";
import { passwordStrength } from "../src/lib/password";

describe("toApiError", () => {
  it("reads the kit error body", () => {
    const e = toApiError(400, {
      error: {
        code: "invalid_argument",
        message: "the request is invalid",
        violations: [{ path: "password", rule: "minlen", message: "must be at least 10 characters" }],
      },
    });
    expect(e).toBeInstanceOf(ApiError);
    expect(e.status).toBe(400);
    expect(e.code).toBe("invalid_argument");
    expect(e.violation("password")).toBe("must be at least 10 characters");
  });

  it("survives a body that is not an error", () => {
    expect(toApiError(502, "<html>").code).toBe("internal");
    expect(toApiError(418, null).code).toBe("bad_response");
  });
});

describe("errorMessage", () => {
  it("maps the login refusals", () => {
    expect(errorMessage(new ApiError(401, "invalid_credentials", "x"))).toMatch(/don’t match/);
    expect(errorMessage(new ApiError(403, "email_unverified", "x"))).toMatch(/Confirm your email/);
    expect(errorMessage(new ApiError(403, "account_locked", "x"))).toMatch(/locked/);
  });

  it("keeps the backend's wording for conflicts", () => {
    expect(errorMessage(new ApiError(409, "conflict", "you are already contacts"))).toBe("You are already contacts");
  });

  it("prefers the single violation of an invalid request", () => {
    const e = new ApiError(400, "invalid_argument", "the request is invalid", [
      { path: "email", rule: "email", message: "must be an email address" },
    ]);
    expect(errorMessage(e)).toBe("Must be an email address");
    expect(fieldErrors(e)).toEqual({ email: "Must be an email address" });
  });

  it("explains the network and unknown failures", () => {
    expect(errorMessage(new ApiError(0, NETWORK_ERROR, ""))).toMatch(/reach the server/);
    expect(errorMessage(new ApiError(500, "internal", "internal error"))).toMatch(/our side/);
    expect(errorMessage(new TypeError("boom"), "Fallback")).toBe("Fallback");
  });

  it("fieldErrors keys nested paths by their first member", () => {
    const e = new ApiError(400, "invalid_argument", "", [{ path: "items[2].zip", rule: "required", message: "is required" }]);
    expect(fieldErrors(e)).toEqual({ items: "Is required" });
  });

  it("isUnauthenticated", () => {
    expect(isUnauthenticated(new ApiError(401, "unauthenticated", ""))).toBe(true);
    expect(isUnauthenticated(new ApiError(401, "invalid_credentials", ""))).toBe(false);
    expect(isUnauthenticated(new ApiError(403, "permission_denied", ""))).toBe(false);
  });
});

describe("passwordStrength", () => {
  it("follows the backend's 10–128 rule", () => {
    expect(passwordStrength("").ok).toBe(false);
    expect(passwordStrength("short").ok).toBe(false);
    expect(passwordStrength("short").label).toBe("5 more characters");
    expect(passwordStrength("x".repeat(129)).ok).toBe(false);
    expect(passwordStrength("abcdefghij").ok).toBe(true);
  });

  it("rewards length and variety, punishes the obvious", () => {
    expect(passwordStrength("password1234").score).toBeLessThan(passwordStrength("Tangerine-Otter-42").score);
    expect(passwordStrength("Tangerine-Otter-42").label).toBe("Strong");
  });
});
