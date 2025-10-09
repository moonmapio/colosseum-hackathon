'use client';

import * as React from 'react';
import { useController } from 'react-hook-form';
import { Button } from '@modules/ui/button';
import { Input } from '@modules/ui/input';
import { Label } from '@modules/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@modules/ui/popover';
import { CalendarClock, ChevronDown, Clock4 } from 'lucide-react';
import { Calendar } from '@modules/ui/calendar';
import { useT } from 'src/i18n/client';

type Props = { name?: string; label?: string };

function parseISOOrUndefined(v?: string) {
	if (!v) {
		const now = new Date();
		now.setHours(now.getHours() + 2);
		return now;
	}
	const d = new Date(v);
	return isNaN(d.getTime()) ? undefined : d;
}

function toISO(date?: Date) {
	return date ? new Date(date.getTime()).toISOString() : undefined;
}

export function LaunchDateField({ name = 'launchDate', label = 'Launch date' }: Props) {
	const language = 'en';

	const { field, fieldState } = useController({ name });
	const [open, setOpen] = React.useState(false);
	const dateObj = parseISOOrUndefined(field.value);
	const [date, setDate] = React.useState<Date | undefined>(dateObj);
	const [time, setTime] = React.useState(() => {
		if (!dateObj) return '08:00';
		const hh = String(dateObj.getHours()).padStart(2, '0');
		const mm = String(dateObj.getMinutes()).padStart(2, '0');
		return `${hh}:${mm}`;
	});

	React.useEffect(() => {
		if (!date) {
			field.onChange(undefined);
			return;
		}
		const [hh, mm] = time.split(':');
		const d = new Date(date);
		d.setHours(Number(hh) || 0, Number(mm) || 0, 0, 0); // segundos = 0
		field.onChange(toISO(d));
	}, [date, field, time]);

	const buttonText = date
		? date.toLocaleDateString(language, { day: '2-digit', month: '2-digit', year: 'numeric' })
		: language.startsWith('es')
			? 'Seleccionar fecha'
			: 'Select date';
	return (
		<div className="flex gap-4 items-end">
			<div className="flex flex-col gap-2 w-full">
				<Label htmlFor="launch-date" className="px-1 flex items-start gap-x-0">
					<div className="mr-1">
						<CalendarClock className="size-3.5" />
					</div>
					{label}
				</Label>
				<Popover open={open} onOpenChange={setOpen} modal={true}>
					<PopoverTrigger asChild>
						<Button variant="outline" id="launch-date" className="w-full justify-between font-normal text-xs">
							{buttonText}
							<ChevronDown className="size-4" />
						</Button>
					</PopoverTrigger>
					<PopoverContent
						className="w-auto overflow-hidden p-0"
						align="end"
						side="left"
						onEscapeKeyDown={(e) => {
							e.preventDefault();
							setOpen(false);
						}}
					>
						<Calendar
							mode="single"
							selected={date}
							captionLayout="dropdown"
							className="[&_[role=gridcell].bg-accent]:bg-sidebar-primary [&_[role=gridcell].bg-accent]:text-sidebar-primary-foreground [&_[role=gridcell]]:w-[33px]"
							onSelect={(d) => {
								setDate(d);
								setOpen(false);
							}}
						/>
					</PopoverContent>
				</Popover>
				{fieldState.error?.message ? <p className="text-xs text-destructive">{fieldState.error.message}</p> : null}
			</div>

			<div className="flex flex-col gap-2 w-full">
				<Label htmlFor="launch-time" className="px-1 flex items-start gap-x-0">
					<div className="mr-1">
						<Clock4 className="size-3.5" />
					</div>
					Launch Time
				</Label>
				<Input
					type="time"
					id="launch-time"
					step={60}
					value={time}
					onChange={(e) => setTime(e.target.value)}
					className="bg-background appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none w-36 !text-xs"
				/>
			</div>
		</div>
	);
}
