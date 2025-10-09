'use client';

import { formatYMD, useSelectedDay } from '@modules/client/useSelectedDay';
import { useAppStore } from '@modules/stores/main/app-hooks';
import { useSelectMintWithUrl } from '@modules/stores/main/ephemeral-hooks';
import { Button } from '@modules/ui/button';
import { Calendar } from '@modules/ui/calendar';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@modules/ui/card';
import { SidebarSeparator, useSidebar } from '@modules/ui/sidebar';
import * as React from 'react';

type Props = {
	dateISO: string;
};

export function CalendarCard(props: Props) {
	const { state, isMobile } = useSidebar();
	const show = state !== 'collapsed' || isMobile;

	const selectMint = useSelectMintWithUrl();
	const setSelectedMenu = useAppStore((s) => s.setSelectedMenu);

	const { dateISO } = props;
	const initialDate = new Date(dateISO + 'T00:00:00Z');

	const { setDay } = useSelectedDay();

	// const [pending, startTransition] = React.useTransition();
	const [date, setDate] = React.useState<Date | undefined>(initialDate);
	const [month, setMonth] = React.useState<Date | undefined>(initialDate);

	const onSelect = React.useCallback(
		(d?: Date) => {
			if (!d) return;
			setDate(d);
			const iso = formatYMD(d);
			setDay(iso);
			// startTransition(() => (iso));
			setSelectedMenu('PROJECTS');
			selectMint(undefined);
		},
		[selectMint, setDay, setSelectedMenu],
	);

	const onToday = React.useCallback(() => {
		const now = new Date();
		setMonth(now);
		onSelect(now);
	}, [onSelect]);

	return (
		<Card className="border-none shadow-none bg-inherit py-2 gap-y-2" hidden={!show}>
			<CardHeader className="px-4 pr-2 gap-y-0">
				<CardTitle className="text-xs mb-0 pb-0">Releases</CardTitle>
				<CardDescription className="text-xs pt-0 mt-0">Find a date</CardDescription>
				<CardAction>
					<Button
						size="sm"
						variant="outline"
						className="hover:cursor-pointer"
						onClick={onToday}
						// disabled={pending}
					>
						Today
					</Button>
				</CardAction>
			</CardHeader>
			<CardContent className="p-0">
				<Calendar
					mode="single"
					month={month}
					onMonthChange={setMonth}
					selected={date}
					onSelect={onSelect}
					className="bg-transparent"
					// disabled={pending}
				/>
			</CardContent>
		</Card>
	);
}

export const CalendarCardSeparator = () => {
	const { state, isMobile } = useSidebar();
	const show = state !== 'collapsed' || isMobile;
	return <SidebarSeparator className="mx-0" hidden={!show} />;
};
