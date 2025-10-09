import { SidebarMenu, SidebarMenuItem } from '@modules/ui/sidebar';
import Image from 'next/image';

export const MoonMapLabel = () => {
	return (
		<SidebarMenu>
			<SidebarMenuItem className="flex flex-row">
				<Image src={'/logo.svg'} alt={'MoonMap'} width={32} height={32} className="ml-2" />
				<h1 className="text-3xl font-semibold ml-2">MoonMap</h1>
			</SidebarMenuItem>
		</SidebarMenu>
	);
};
