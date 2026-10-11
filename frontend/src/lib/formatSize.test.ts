import { describe, expect, it } from "vitest";
import { formatSize } from "./formatSize";

describe("formatSize", () => {
  it("picks the unit by size", () => {
    expect(formatSize(512)).toBe("512 B");
    expect(formatSize(2048)).toBe("2.0 KB");
    expect(formatSize(1572864)).toBe("1.5 MB");
  });
});
