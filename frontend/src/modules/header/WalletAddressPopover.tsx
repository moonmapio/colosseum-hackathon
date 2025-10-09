'use client';

import { getWalletSecret } from '@modules/server/wallet';
import { useAppStore } from '@modules/stores/main/app-hooks';
import { AppStoreState } from '@modules/stores/main/app-store';
import { useEphemeralAppStore } from '@modules/stores/main/ephemeral-hooks';
import { Button } from '@modules/ui/button';
import { Popover, PopoverContent, PopoverTrigger } from '@modules/ui/popover';
import { Copy, Download, Wallet } from 'lucide-react';
import { useCallback, useState } from 'react';

export function WalletAddressPopover() {
	const [open, setOpen] = useState(false);

	const selector = useCallback((state: AppStoreState) => {
		return {
			download: state.wallet?.downloaded,
			walletAddress: state.wallet?.walletAddress,
			createdAt: state.wallet?.createdAt,
			balance: state.balance,
		};
	}, []);

	const { walletAddress, createdAt, balance, download } = useAppStore(selector);
	const setWalletDownloaded = useAppStore((state) => state.setWalletDownloaded);
	const isMobile = useEphemeralAppStore((state) => state.isMobile);

	const onCopy = useCallback(() => navigator.clipboard.writeText(walletAddress ?? ''), [walletAddress]);
	const downloadSecret = useCallback(async () => {
		if (!walletAddress) return;
		const secret = await getWalletSecret(walletAddress);
		setWalletDownloaded();
		const blob = new Blob([secret], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `moonmap-wallet-${walletAddress}.txt`;
		a.click();
		URL.revokeObjectURL(url);
	}, [walletAddress, setWalletDownloaded]);

	if (!walletAddress || !createdAt) return null;

	const shortAddr = walletAddress.slice(0, 4) + '...' + walletAddress.slice(-4);

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<Button
					variant="outline"
					data-mobile={isMobile}
					className="hover:cursor-pointer data-[mobile=true]:text-[10px] data-[mobile=true]:!p-1.5"
				>
					{isMobile ? '' : shortAddr}
					<Wallet size={26} data-mobile={isMobile} className="text-primary data-[mobile=true]:text-[10px]" strokeWidth={2.2} />
				</Button>
			</PopoverTrigger>
			<PopoverContent className="w-80 p-4 group" data-mobile={isMobile} side="bottom" align="end">
				{!download && (
					<div className="bg-muted rounded-md p-2">
						<p className="text-xs text-accent-foreground font-mono group-data-[mobile=true]:text-[10px]">
							Download and store your secret key safely. Without it, you can’t recover your wallet — whoever has it controls your funds.
						</p>
					</div>
				)}
				<div className="space-y-1 text-sm mt-2">
					<div>
						<span className="font-medium group-data-[mobile=true]:text-xs">Wallet Address:</span>
						<div className="flex items-center gap-2 break-all">
							<code className="text-xs">{walletAddress}</code>
							<Button size="icon" variant="ghost" onClick={onCopy}>
								<Copy className="h-4 w-4" />
							</Button>
						</div>
					</div>

					<div className="flex flex-row group-data-[mobile=true]:text-xs">
						<span className="font-medium mr-2">Balance:</span>
						<div>
							<span className="font-bold">{balance.toFixed(4)} SOL</span>
						</div>
					</div>

					<div className="flex flex-row group-data-[mobile=true]:text-xs">
						<span className="font-medium mr-2">Created At:</span>
						<div>
							{new Date(createdAt).toLocaleString('default', {
								month: 'long',
								day: '2-digit',
								year: 'numeric',
								hour: 'numeric',
								minute: 'numeric',
							})}
						</div>
					</div>

					{!download && (
						<div className="pt-2">
							<Button variant="default" onClick={downloadSecret} className="w-full flex gap-2 hover:cursor-pointer">
								<Download className="h-4 w-4" /> Download Secret
							</Button>
						</div>
					)}
				</div>
			</PopoverContent>
		</Popover>
	);
}
