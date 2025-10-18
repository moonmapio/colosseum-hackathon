'use client';

import { setSearchQuery } from '@modules/server/setSearchCookie';
import { useAppStore } from '@modules/stores/main/app-hooks';
import { useEphemeralAppStore } from '@modules/stores/main/ephemeral-hooks';
import { Input } from '@modules/ui/input';
import { useDebounce } from '@uidotdev/usehooks';
import { ChangeEventHandler, useCallback, useEffect, useState } from 'react';

type Props = {
	searchQuery: string;
};

export function SearchComponent(props: Props) {
	const { searchQuery } = props;
	const selectedMenu = useAppStore((state) => state.selectedMenu);
	const show = selectedMenu === 'PROJECTS';
	const isMobile = useEphemeralAppStore((state) => state.isMobile);

	const [searchTerm, setSearchTerm] = useState(searchQuery !== '*' ? searchQuery : '');
	const debouncedSearchTerm = useDebounce(searchTerm, 500);

	const onChange: ChangeEventHandler<HTMLInputElement> = useCallback((event) => {
		setSearchTerm(event.target.value);
	}, []);

	useEffect(() => {
		let result = '*';
		if (debouncedSearchTerm) {
			result = debouncedSearchTerm;
		}
		if (searchQuery !== result) {
			setSearchQuery(result);
		}
	}, [debouncedSearchTerm, searchQuery]);

	if (!show || isMobile) return null;

	return (
		<div className="w-full max-w-sm">
			<Input type="text" placeholder="Search ..." value={searchTerm} onChange={onChange} />
		</div>
	);
}

// 'use client';

// import * as React from 'react';
// import { Calculator, Calendar, CreditCard, Settings, Smile, User } from 'lucide-react';

// import {
// 	CommandDialog,
// 	CommandEmpty,
// 	CommandGroup,
// 	CommandInput,
// 	CommandItem,
// 	CommandList,
// 	CommandSeparator,
// 	CommandShortcut,
// } from '@modules/ui/command';

// export function SearchCommand() {
// 	const [open, setOpen] = React.useState(false);

// 	React.useEffect(() => {
// 		const down = (e: KeyboardEvent) => {
// 			if (e.key === 'j' && (e.metaKey || e.ctrlKey)) {
// 				e.preventDefault();
// 				setOpen((open) => !open);
// 			}
// 		};

// 		document.addEventListener('keydown', down);
// 		return () => document.removeEventListener('keydown', down);
// 	}, []);

// 	return (
// 		<>
// 			<p className="text-muted-foreground text-sm">
// 				Press{' '}
// 				<kbd className="bg-muted text-muted-foreground pointer-events-none inline-flex h-5 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium opacity-100 select-none">
// 					<span className="text-xs">⌘</span>J
// 				</kbd>
// 			</p>
// 			<CommandDialog open={open} onOpenChange={setOpen}>
// 				<CommandInput placeholder="Type a command or search..." />
// 				<CommandList>
// 					<CommandEmpty>No results found.</CommandEmpty>
// 					<CommandGroup heading="Suggestions">
// 						<CommandItem>
// 							<Calendar />
// 							<span>Calendar</span>
// 						</CommandItem>
// 						<CommandItem>
// 							<Smile />
// 							<span>Search Emoji</span>
// 						</CommandItem>
// 						<CommandItem>
// 							<Calculator />
// 							<span>Calculator</span>
// 						</CommandItem>
// 					</CommandGroup>
// 					<CommandSeparator />
// 					<CommandGroup heading="Settings">
// 						<CommandItem>
// 							<User />
// 							<span>Profile</span>
// 							<CommandShortcut>⌘P</CommandShortcut>
// 						</CommandItem>
// 						<CommandItem>
// 							<CreditCard />
// 							<span>Billing</span>
// 							<CommandShortcut>⌘B</CommandShortcut>
// 						</CommandItem>
// 						<CommandItem>
// 							<Settings />
// 							<span>Settings</span>
// 							<CommandShortcut>⌘S</CommandShortcut>
// 						</CommandItem>
// 					</CommandGroup>
// 				</CommandList>
// 			</CommandDialog>
// 		</>
// 	);
// }

// 'use client';

// import { KeyboardEventHandler, useCallback, useState } from 'react';
// import { Command, CommandDialog, CommandInput } from '@modules/ui/command'; // si usas shadcn/ui
// import { setSearchQuery } from '@modules/server/setSearchCookie';
// import { Button } from '@modules/ui/button';

// export function SearchCommand() {
// 	const [open, setOpen] = useState(false);

// 	const onKeyDown: KeyboardEventHandler<HTMLInputElement> = useCallback(async (event) => {
// 		if (event.key === 'Enter') {
// 			const q = (event.target as HTMLInputElement).value;
// 			await setSearchQuery(q);
// 			setOpen(false);
// 		}
// 	}, []);

// 	return (
// 		<Command>
//             <CommandInput  placeholder="Type a command or search...">

//             </CommandInput>

// 			<Button onClick={() => setOpen(true)} className="text-sm text-muted-foreground hover:text-foreground">
// 				⌘K Search
// 			</Button>
// 			<CommandDialog open={open} onOpenChange={setOpen}>
// 				<Command>
// 					<CommandInput placeholder="Search..." onKeyDown={onKeyDown} />
// 				</Command>
// 			</CommandDialog>
//         </Command>
// 	);
// }
