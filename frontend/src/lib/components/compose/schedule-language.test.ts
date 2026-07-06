import { describe, expect, it } from 'vitest';
import { parseNaturalScheduleInput } from './schedule-language';

const base = new Date(2026, 6, 6, 10, 15, 0, 0);

describe('parseNaturalScheduleInput', () => {
	it('parses tomorrow with a meridiem time', () => {
		const parsed = parseNaturalScheduleInput('tomorrow at 9am', base);

		expect(parsed?.date.toString()).toBe('2026-07-07');
		expect(parsed?.time).toBe('09:00');
	});

	it('parses relative hour offsets', () => {
		const parsed = parseNaturalScheduleInput('in 3 hours', base);

		expect(parsed?.date.toString()).toBe('2026-07-06');
		expect(parsed?.time).toBe('13:15');
	});

	it('rolls a plain past time to tomorrow', () => {
		const parsed = parseNaturalScheduleInput('9:30', base);

		expect(parsed?.date.toString()).toBe('2026-07-07');
		expect(parsed?.time).toBe('09:30');
	});

	it('parses next weekdays', () => {
		const parsed = parseNaturalScheduleInput('next monday at noon', base);

		expect(parsed?.date.toString()).toBe('2026-07-13');
		expect(parsed?.time).toBe('12:00');
	});

	it('parses month names', () => {
		const parsed = parseNaturalScheduleInput('July 10 4:45pm', base);

		expect(parsed?.date.toString()).toBe('2026-07-10');
		expect(parsed?.time).toBe('16:45');
	});
});
