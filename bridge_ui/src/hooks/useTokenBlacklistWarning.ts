import {
  ChainId,
  CHAIN_ID_ETH,
  CHAIN_ID_SOLANA,
} from "@certusone/wormhole-sdk";
import { useMemo } from "react";
import {
  ETH_TOKENS_THAT_EXIST_ELSEWHERE,
  SOLANA_TOKENS_THAT_EXIST_ELSEWHERE,
} from "../utils/consts";

export default function useTokenBlacklistWarning(
  chainId: ChainId,
  tokenAddress: string | undefined,
  symbol: string | undefined
) {
  return useMemo(
    () =>
      tokenAddress &&
      ((chainId === CHAIN_ID_SOLANA &&
        SOLANA_TOKENS_THAT_EXIST_ELSEWHERE.includes(tokenAddress)) ||
        (chainId === CHAIN_ID_ETH &&
          ETH_TOKENS_THAT_EXIST_ELSEWHERE.includes(tokenAddress)))
        ? `Bridging ${
            symbol ? symbol : "the token"
          } via Wormhole will not produce native ${
            symbol ? symbol : "assets"
          }. It will produce a wrapped version which might have no liquidity or utility on the target chain.`
        : undefined,
    [chainId, tokenAddress, symbol]
  );
}
