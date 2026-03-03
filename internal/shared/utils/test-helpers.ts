import { expect } from "@jest/globals";

/**
 * Assert common cleaner field values match test expectations
 */
export function assertCleanerFields(
  cleaner: any,
  expectedVerbose: boolean,
  expectedDryRun: boolean,
): void {
  expect(cleaner.verbose).toBe(expectedVerbose);
  expect(cleaner.dryRun).toBe(expectedDryRun);
}
