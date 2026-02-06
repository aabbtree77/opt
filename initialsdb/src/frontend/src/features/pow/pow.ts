// src/features/pow/pow.ts
export type PowChallenge = {
  challenge: string;
  difficulty: number;
  ttl_secs: number;
  token: string;
};

export async function getChallenge(): Promise<PowChallenge> {
  const r = await fetch("/pow/challenge", { credentials: "same-origin" });
  if (!r.ok) throw new Error("challenge unavailable");
  return r.json();
}

function hasLeadingZeroBits(hash: Uint8Array, difficulty: number) {
  let bits = 0;
  for (const b of hash) {
    for (let i = 7; i >= 0; i--) {
      if (bits === difficulty) return true;
      if ((b >> i) & 1) return false;
      bits++;
    }
  }
  return bits >= difficulty;
}

export async function solvePoW(
  challenge: string,
  difficulty: number,
  ttlSecs: number,
  onProgress?: (tries: number, remaining: number) => void
): Promise<string> {
  const enc = new TextEncoder();
  const chBytes = Uint8Array.from(atob(challenge), (c) => c.charCodeAt(0));

  const deadline = Date.now() + ttlSecs * 1000;
  let nonce = 0;

  while (true) {
    if (Date.now() > deadline) {
      throw new Error("PoW expired");
    }

    if (nonce % 5000 === 0 && onProgress) {
      const remaining = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
      onProgress(nonce, remaining);
      await new Promise((r) => setTimeout(r, 0));
    }

    const nonceStr = String(nonce);
    const data = new Uint8Array(chBytes.length + nonceStr.length);
    data.set(chBytes);
    data.set(enc.encode(nonceStr), chBytes.length);

    const hash = new Uint8Array(await crypto.subtle.digest("SHA-256", data));

    if (hasLeadingZeroBits(hash, difficulty)) {
      return nonceStr;
    }

    nonce++;
  }
}
