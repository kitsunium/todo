import { describe, expect, it } from "vitest";
import { ApiError, NETWORK_ERROR, toApiError } from "../src/api/client";
import { errorMessage, fieldErrors, isUnauthenticated } from "../src/api/errors";
import { passwordStrength } from "../src/lib/password";

const fr = { locale: "fr" } as const;
const en = { locale: "en" } as const;

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
  it("maps the login refusals, in French first", () => {
    expect(errorMessage(new ApiError(401, "invalid_credentials", "x"), fr)).toBe("Cet e\u2011mail et ce mot de passe ne correspondent pas.");
    expect(errorMessage(new ApiError(403, "email_unverified", "x"), fr)).toMatch(/^Confirmez d’abord votre adresse e\u2011mail/);
    expect(errorMessage(new ApiError(403, "account_locked", "x"), fr)).toMatch(/verrouillé pendant 15 minutes/);
    expect(errorMessage(new ApiError(401, "invalid_credentials", "x"), en)).toMatch(/don’t match/);
    expect(errorMessage(new ApiError(403, "email_unverified", "x"), en)).toMatch(/Confirm your email/);
    expect(errorMessage(new ApiError(403, "account_locked", "x"), en)).toMatch(/locked/);
  });

  it.each([
    ["invalid_token", /lien n’est plus valide/, /invalid or has expired/],
    ["rate_limited", /quelques secondes/, /Slow down/],
    ["unauthenticated", /session a expiré/, /session has ended/],
    ["permission_denied", /autorisation/, /permission/],
    ["not_found", /plus là/, /isn’t here/],
  ])("%s never shows the server's English", (code, frText, enText) => {
    const e = new ApiError(400, code, "the server's own english words");
    expect(errorMessage(e, fr)).toMatch(frText);
    expect(errorMessage(e, en)).toMatch(enText);
    expect(errorMessage(e, fr)).not.toMatch(/english words/);
  });

  it("says what a known conflict collided with", () => {
    expect(errorMessage(new ApiError(409, "conflict", "you are already contacts"), fr)).toBe("Vous êtes déjà en contact.");
    expect(errorMessage(new ApiError(409, "conflict", "you are already contacts"), en)).toBe("You are already contacts.");
    expect(errorMessage(new ApiError(409, "conflict", "this person is already a member"), fr)).toBe(
      "Cette personne fait déjà partie du groupe.",
    );
    expect(errorMessage(new ApiError(409, "conflict", "this task cannot be completed now"), fr)).toMatch(/Cette tâche a changé/);
  });

  it("an unknown conflict gets the generic sentence, not the server's", () => {
    expect(errorMessage(new ApiError(409, "conflict", "something new happened"), fr)).toMatch(/^Ce n’est plus possible/);
    expect(errorMessage(new ApiError(409, "conflict", "something new happened"), en)).toMatch(/^That’s no longer possible/);
  });

  it("says the single violation of an invalid request with its field", () => {
    const e = new ApiError(400, "invalid_argument", "the request is invalid", [
      { path: "email", rule: "email", message: "must be an email address" },
    ]);
    expect(errorMessage(e, fr)).toBe("E\u2011mail\u00a0: adresse e\u2011mail invalide.");
    expect(errorMessage(e, en)).toBe("Email: not a valid email address.");
    expect(fieldErrors(e, "fr")).toEqual({ email: "Adresse e\u2011mail invalide." });
    expect(fieldErrors(e, "en")).toEqual({ email: "Not a valid email address." });
  });

  it("reads a rule's limit from the server's message", () => {
    const max = new ApiError(400, "invalid_argument", "", [{ path: "title", rule: "maxlen", message: "must be at most 200 characters long" }]);
    expect(errorMessage(max, fr)).toBe("Titre\u00a0: 200 caractères maximum.");
    expect(errorMessage(max, en)).toBe("Title: at most 200 characters.");
    const min = new ApiError(400, "invalid_argument", "", [{ path: "password", rule: "minlen", message: "must be at least 10 characters long" }]);
    expect(fieldErrors(min, "fr")).toEqual({ password: "10 caractères minimum." });
    const req = new ApiError(400, "invalid_argument", "", [{ path: "name", rule: "required", message: "is required" }]);
    expect(errorMessage(req, fr)).toBe("Nom\u00a0: obligatoire.");
    expect(errorMessage(req, en)).toBe("Name: required.");
    const loc = new ApiError(400, "invalid_argument", "", [{ path: "locale", rule: "one_of", message: "must be one of: fr, en" }]);
    expect(errorMessage(loc, fr)).toBe("Langue\u00a0: valeur non autorisée.");
    const self = new ApiError(400, "invalid_argument", "", [{ path: "email", rule: "self", message: "is your own address" }]);
    expect(fieldErrors(self, "fr")).toEqual({ email: "C’est votre propre adresse." });
  });

  it("several violations: one general sentence", () => {
    const e = new ApiError(400, "invalid_argument", "the request is invalid", [
      { path: "name", rule: "required", message: "is required" },
      { path: "email", rule: "email", message: "must be an email address" },
    ]);
    expect(errorMessage(e, fr)).toBe("Certaines informations ne sont pas valides.");
    expect(fieldErrors(e, "fr")).toEqual({ name: "Obligatoire.", email: "Adresse e\u2011mail invalide." });
  });

  it("explains the network and unknown failures", () => {
    expect(errorMessage(new ApiError(0, NETWORK_ERROR, ""), fr)).toMatch(/joindre le serveur/);
    expect(errorMessage(new ApiError(0, NETWORK_ERROR, ""), en)).toMatch(/reach the server/);
    expect(errorMessage(new ApiError(500, "internal", "internal error"), en)).toMatch(/our side/);
    expect(errorMessage(new TypeError("boom"), { fallback: "Fallback" })).toBe("Fallback");
    expect(errorMessage(new TypeError("boom"), fr)).toBe("Un problème est survenu. Réessayez.");
  });

  it("keeps the server's words only for a code it doesn't know", () => {
    expect(errorMessage(new ApiError(418, "teapot", "i am a teapot"), fr)).toBe("I am a teapot.");
  });

  it("fieldErrors keys nested paths by their first member", () => {
    const e = new ApiError(400, "invalid_argument", "", [{ path: "items[2].zip", rule: "required", message: "is required" }]);
    expect(fieldErrors(e, "en")).toEqual({ items: "Required." });
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
    expect(passwordStrength("short", "en").label).toBe("5 more characters");
    expect(passwordStrength("short", "fr").label).toBe("Encore 5 caractères");
    expect(passwordStrength("123456789", "fr").label).toBe("Encore 1 caractère");
    expect(passwordStrength("x".repeat(129)).ok).toBe(false);
    expect(passwordStrength("abcdefghij").ok).toBe(true);
  });

  it("rewards length and variety, punishes the obvious", () => {
    expect(passwordStrength("password1234").score).toBeLessThan(passwordStrength("Tangerine-Otter-42").score);
    expect(passwordStrength("Tangerine-Otter-42", "en").label).toBe("Strong");
    expect(passwordStrength("Tangerine-Otter-42", "fr").label).toBe("Robuste");
  });
});
