'use server';

import { ModeToggle } from '@modules/header/ModeToggle';
import { Separator } from '@modules/ui/separator';
import { SidebarTrigger } from '@modules/ui/sidebar';
import { SearchComponent } from './SearchComponent';
import { SolanaAndBtcPrice } from './SolanaAndBtcPrice';
import { WalletAddressPopover } from './WalletAddressPopover';
import { HeaderLogo } from './HeaderLogo';

type Props = {
	// language: string; dateISO: string;
	searchQuery: string;
};

export const AppHeader = async (props: Props) => {
	const { searchQuery } = props;

	return (
		// <header className="bg-background sticky top-0 flex h-16 shrink-0 items-center gap-2 border-b px-4 z-50">
		<header className="flex h-16 shrink-0 items-center gap-2">
			<div className="flex items-center gap-2 px-4 w-full">
				<SidebarTrigger className="-ml-1" />
				<Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-8" />
				<HeaderLogo />
				<SearchComponent searchQuery={searchQuery} />
				<div className="ml-auto">
					<SolanaAndBtcPrice />
				</div>
				<ModeToggle />
				<WalletAddressPopover />
			</div>
		</header>
	);
};
