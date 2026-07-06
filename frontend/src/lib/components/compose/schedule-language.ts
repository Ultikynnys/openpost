import { CalendarDate } from '@internationalized/date';

export interface ParsedScheduleInput {
	date: CalendarDate;
	time: string;
}

const WEEKDAY_INDEXES: Record<string, number> = {
	sunday: 0,
	sun: 0,
	monday: 1,
	mon: 1,
	tuesday: 2,
	tue: 2,
	tues: 2,
	wednesday: 3,
	wed: 3,
	thursday: 4,
	thu: 4,
	thurs: 4,
	friday: 5,
	fri: 5,
	saturday: 6,
	sat: 6
};

const MONTH_INDEXES: Record<string, number> = {
	january: 0,
	jan: 0,
	february: 1,
	feb: 1,
	march: 2,
	mar: 2,
	april: 3,
	apr: 3,
	may: 4,
	june: 5,
	jun: 5,
	july: 6,
	jul: 6,
	august: 7,
	aug: 7,
	september: 8,
	sep: 8,
	sept: 8,
	october: 9,
	oct: 9,
	november: 10,
	nov: 10,
	december: 11,
	dec: 11
};

function toCalendarDate(date: Date): CalendarDate {
	return new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate());
}

function toTime(date: Date): string {
	return `${date.getHours().toString().padStart(2, '0')}:${date
		.getMinutes()
		.toString()
		.padStart(2, '0')}`;
}

function parsed(date: Date): ParsedScheduleInput {
	return {
		date: toCalendarDate(date),
		time: toTime(date)
	};
}

function parseTime(input: string): { hour: number; minute: number } | null {
	if (/\bnoon\b/.test(input)) return { hour: 12, minute: 0 };
	if (/\bmidnight\b/.test(input)) return { hour: 0, minute: 0 };

	const matches = Array.from(input.matchAll(/\b(at\s*)?(\d{1,2})(?::(\d{2}))?\s*(am|pm)?\b/g));
	const explicitMatches = matches.filter((match) => !!match[1] || !!match[3] || !!match[4]);
	const match =
		explicitMatches.length > 0
			? explicitMatches[explicitMatches.length - 1]
			: matches[matches.length - 1];
	if (!match) return null;

	let hour = Number.parseInt(match[2], 10);
	const minute = match[3] ? Number.parseInt(match[3], 10) : 0;
	const meridiem = match[4];

	if (!Number.isFinite(hour) || !Number.isFinite(minute) || minute < 0 || minute > 59) {
		return null;
	}
	if (meridiem === 'am') {
		if (hour === 12) hour = 0;
	} else if (meridiem === 'pm') {
		if (hour < 12) hour += 12;
	}
	if (hour < 0 || hour > 23) return null;
	return { hour, minute };
}

function applyTime(date: Date, time: { hour: number; minute: number } | null, fallbackHour = 9) {
	const next = new Date(date);
	next.setHours(time?.hour ?? fallbackHour, time?.minute ?? 0, 0, 0);
	return next;
}

function nextWeekday(base: Date, targetDay: number, forceNext: boolean, timeInput: string): Date {
	const time = parseTime(timeInput);
	let daysUntil = (targetDay - base.getDay() + 7) % 7;
	if (forceNext || daysUntil === 0) {
		const candidate = applyTime(base, time);
		if (forceNext || candidate.getTime() <= base.getTime())
			daysUntil = daysUntil === 0 ? 7 : daysUntil;
	}
	const next = new Date(base);
	next.setDate(base.getDate() + daysUntil);
	return applyTime(next, time);
}

export function parseNaturalScheduleInput(
	input: string,
	now = new Date()
): ParsedScheduleInput | null {
	const normalized = input.trim().toLowerCase().replace(/[,]+/g, ' ').replace(/\s+/g, ' ');
	if (!normalized) return null;

	const relativeMatch = normalized.match(
		/^in\s+(\d+)\s*(minutes?|mins?|m|hours?|hrs?|h|days?|d|weeks?|w)$/
	);
	if (relativeMatch) {
		const amount = Number.parseInt(relativeMatch[1], 10);
		const unit = relativeMatch[2];
		const next = new Date(now);
		if (unit.startsWith('m')) next.setMinutes(next.getMinutes() + amount);
		else if (unit.startsWith('h')) next.setHours(next.getHours() + amount);
		else if (unit.startsWith('d')) next.setDate(next.getDate() + amount);
		else next.setDate(next.getDate() + amount * 7);
		next.setSeconds(0, 0);
		return parsed(next);
	}

	const tomorrowMatch = /\btomorrow\b/.test(normalized);
	const todayMatch = /\btoday\b/.test(normalized);
	const tonightMatch = /\btonight\b/.test(normalized);
	if (tomorrowMatch || todayMatch || tonightMatch) {
		const next = new Date(now);
		if (tomorrowMatch) next.setDate(now.getDate() + 1);
		const time = parseTime(normalized);
		const fallbackHour = tonightMatch ? 20 : 9;
		const candidate = applyTime(next, time, fallbackHour);
		if (!tomorrowMatch && candidate.getTime() <= now.getTime()) {
			candidate.setDate(candidate.getDate() + 1);
		}
		return parsed(candidate);
	}

	const weekdayMatch = normalized.match(
		/\b(next\s+)?(sun(?:day)?|mon(?:day)?|tue(?:s|sday)?|wed(?:nesday)?|thu(?:rs|rsday)?|fri(?:day)?|sat(?:urday)?)\b/
	);
	if (weekdayMatch) {
		const weekday = WEEKDAY_INDEXES[weekdayMatch[2]];
		if (weekday !== undefined) {
			return parsed(
				nextWeekday(now, weekday, !!weekdayMatch[1], normalized.replace(weekdayMatch[0], ''))
			);
		}
	}

	const monthMatch = normalized.match(
		/\b(jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:t|tember)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\s+(\d{1,2})(?:\s+(\d{4}))?\b/
	);
	if (monthMatch) {
		const month = MONTH_INDEXES[monthMatch[1]];
		const day = Number.parseInt(monthMatch[2], 10);
		const explicitYear = monthMatch[3] ? Number.parseInt(monthMatch[3], 10) : null;
		if (month !== undefined && day >= 1 && day <= 31) {
			const candidate = new Date(explicitYear ?? now.getFullYear(), month, day);
			const withTime = applyTime(candidate, parseTime(normalized.replace(monthMatch[0], '')));
			if (!explicitYear && withTime.getTime() <= now.getTime()) {
				withTime.setFullYear(withTime.getFullYear() + 1);
			}
			return parsed(withTime);
		}
	}

	const slashDateMatch = normalized.match(/\b(\d{1,2})\/(\d{1,2})(?:\/(\d{2,4}))?\b/);
	if (slashDateMatch) {
		const month = Number.parseInt(slashDateMatch[1], 10) - 1;
		const day = Number.parseInt(slashDateMatch[2], 10);
		let year = slashDateMatch[3] ? Number.parseInt(slashDateMatch[3], 10) : now.getFullYear();
		if (year < 100) year += 2000;
		if (month >= 0 && month <= 11 && day >= 1 && day <= 31) {
			const candidate = applyTime(
				new Date(year, month, day),
				parseTime(normalized.replace(slashDateMatch[0], ''))
			);
			if (!slashDateMatch[3] && candidate.getTime() <= now.getTime()) {
				candidate.setFullYear(candidate.getFullYear() + 1);
			}
			return parsed(candidate);
		}
	}

	const time = parseTime(normalized);
	if (time) {
		const candidate = applyTime(now, time);
		if (candidate.getTime() <= now.getTime()) {
			candidate.setDate(candidate.getDate() + 1);
		}
		return parsed(candidate);
	}

	return null;
}
