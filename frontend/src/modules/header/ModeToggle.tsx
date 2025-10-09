'use client';

import { Check, Moon, Sun } from 'lucide-react';
import { useTheme } from 'next-themes';

import { useEphemeralAppStore } from '@modules/stores/main/ephemeral-hooks';
import { Button } from '@modules/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@modules/ui/dropdown-menu';

type Props = {
	className?: string;
};

export function ModeToggle(props: Props) {
	const { className } = props;
	const { setTheme, theme } = useTheme();
	const isMobile = useEphemeralAppStore((state) => state.isMobile);

	const menuItemClassName = 'w-full items-center justify-between';

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<Button variant="outline" size="icon" className={className}>
					<Sun className="h-[1.2rem] w-[1.2rem] scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
					<Moon className="absolute h-[1.2rem] w-[1.2rem] scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
					<span className="sr-only">Toggle theme</span>
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end">
				<DropdownMenuItem className={menuItemClassName} onClick={() => setTheme('light')}>
					<p data-mobile={isMobile} className="data-[mobile=true]:text-xs">
						Light
					</p>
					{theme === 'light' && (
						<div>
							<Check />
						</div>
					)}
				</DropdownMenuItem>
				<DropdownMenuItem className={menuItemClassName} onClick={() => setTheme('dark')}>
					<p data-mobile={isMobile} className="data-[mobile=true]:text-xs">
						Dark
					</p>
					{theme === 'dark' && (
						<div>
							<Check />
						</div>
					)}
				</DropdownMenuItem>
				<DropdownMenuItem className={menuItemClassName} onClick={() => setTheme('system')}>
					<p data-mobile={isMobile} className="data-[mobile=true]:text-xs">
						System
					</p>

					{theme === 'system' && (
						<div>
							<Check />
						</div>
					)}
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}
