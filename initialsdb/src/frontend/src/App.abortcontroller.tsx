import { useCallback, useEffect, useRef, useState } from 'react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Toggle } from '@/components/ui/toggle'
import { getChallenge, solvePoW } from '@/features/pow/pow'

// ==================================================
// Types
// ==================================================

interface Listing {
  id: number
  body: string
  created_at: string
}

interface SearchResponse {
  items: Listing[]
  next_cursor?: string
}

type StatusType = 'info' | 'error'

interface StatusMessage {
  id: number
  text: string
  type: StatusType
}

type AppState =
  | { tag: 'idle' }
  | { tag: 'searching' }
  | { tag: 'search'; cursor: string | null }
  | { tag: 'posting' }
  | { tag: 'pow' }

// ==================================================
// Constants
// ==================================================

const HOVER_MESSAGE = 'For sale: baby shoes, never worn.\nErnest@Hemingway.com'

// ==================================================
// API helpers
// ==================================================

async function searchAPI(
  q: string,
  limit: number,
  cursor: string | null,
  signal: AbortSignal,
): Promise<SearchResponse> {
  const params = new URLSearchParams()
  params.set('q', q)
  params.set('limit', String(limit))
  if (cursor) params.set('cursor', cursor)

  const res = await fetch(`/api/listings/search?${params.toString()}`, {
    signal,
  })

  if (!res.ok) throw new Error('search failed')
  return res.json() as Promise<SearchResponse>
}

async function postAPI(
  text: string,
  pow: { challenge: string; nonce: string; token: string },
  signal: AbortSignal,
): Promise<Listing> {
  const res = await fetch('/api/listings/create', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-PoW-Challenge': pow.challenge,
      'X-PoW-Nonce': pow.nonce,
      'X-PoW-Token': pow.token,
    },
    body: JSON.stringify({ text }),
    signal,
  })

  if (!res.ok) throw new Error('post failed')
  return res.json() as Promise<Listing>
}

async function fetchCount(signal?: AbortSignal): Promise<number> {
  const res = await fetch('/api/listings/count', { signal })
  if (!res.ok) throw new Error('count failed')
  const json = await res.json()
  return json.count as number
}

// ==================================================
// Components
// ==================================================

function StatusLine({
  message,
  hoverMessage,
}: {
  message: StatusMessage | null
  hoverMessage: string | null
}) {
  const displayText = hoverMessage ?? message?.text ?? null
  const displayType = hoverMessage ? 'info' : (message?.type ?? 'info')

  return (
    <div className="mt-6 min-h-[3rem] flex items-center justify-center">
      {displayText && (
        <div
          className={[
            'text-xl text-center whitespace-pre-line',
            displayType === 'error' ? 'text-red-600' : 'text-[#9AA1AC]',
          ].join(' ')}
        >
          {displayText}
        </div>
      )}
    </div>
  )
}

function ControlLine(props: {
  query: string
  disabled: boolean
  postOpen: boolean
  onQueryChange(q: string): void
  onSearch(): void
  onTogglePost(): void
  onMouseEnterPost?(): void
  onMouseLeavePost?(): void
}) {
  return (
    <div>
      {' '}
      <div className="flex flex-col md:flex-row items-center gap-3 justify-center">
        <Input
          value={props.query}
          disabled={props.disabled}
          placeholder="👀"
          onChange={(e) => props.onQueryChange(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') props.onSearch()
          }}
          className="
            h-16
            w-[90vw]
            md:w-[22vw]
            bg-[#2A323C]
            border
            border-[#9AA1AC]
            text-[#9AA1AC]
            rounded-lg
            px-4
            !text-2xl 
            font-medium
          "
        />

        <div className="relative flex items-center justify-center w-[90vw] md:w-[6vw]">
          <Toggle
            pressed={props.postOpen}
            disabled={props.disabled}
            onPressedChange={props.onTogglePost}
            onMouseEnter={props.onMouseEnterPost}
            onMouseLeave={props.onMouseLeavePost}
            className="
              h-16
              w-full
              rounded-lg
              border
              border-[#9AA1AC]
              bg-[#2A323C]
              text-[#9AA1AC]
              text-xl
              transition-colors
              duration-200
              flex
              items-center
              justify-center

              hover:bg-[#3A414C]
              hover:text-[#9AA1AC]

              data-[state=on]:bg-[#8FD3E8]
              data-[state=on]:text-[#2A323C]
              data-[state=on]:border-[#8FD3E8]
            "
          >
            Post
          </Toggle>
        </div>
      </div>
    </div>
  )
}

function SearchResults(props: {
  items: Listing[]
  loading: boolean
  hasMore: boolean
  onLoadMore(): void
}) {
  const loaderRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const node = loaderRef.current
    if (!node) return

    const obs = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && props.hasMore && !props.loading) {
        props.onLoadMore()
      }
    })

    obs.observe(node)
    return () => obs.disconnect()
  }, [props.hasMore, props.loading, props.onLoadMore])

  return (
    <div className="mt-8 space-y-4 text-left">
      {props.items.map((l, i) => (
        <div key={l.id}>
          <div className="text-xs text-[#6E7681] mb-1">
            {new Date(l.created_at).toLocaleString()}
          </div>
          <p className="text-[#9AA1AC] whitespace-pre-wrap text-xl">{l.body}</p>
          {i < props.items.length - 1 && (
            <hr className="my-3 border-[#3A414C] w-full" />
          )}
        </div>
      ))}
      {props.hasMore && <div ref={loaderRef} className="h-10" />}
    </div>
  )
}

function PostForm(props: {
  text: string
  disabled: boolean
  powInfo: string | null
  onTextChange(v: string): void
  onSubmit(): void
}) {
  const MAX = 255
  const remaining = Math.max(0, MAX - props.text.length)

  return (
    <div className="mt-6">
      <form
        onSubmit={(e) => {
          e.preventDefault()
          props.onSubmit()
        }}
        className="space-y-2"
      >
        <textarea
          value={props.text}
          maxLength={MAX}
          onChange={(e) => props.onTextChange(e.target.value.slice(0, MAX))}
          disabled={props.disabled}
          placeholder="Write your post…"
          className="
    w-full
    h-32
    bg-[#2A323C]
    rounded-md
    p-2
    text-xl
    text-[#9AA1AC]
    border
    border-[#9AA1AC]
    focus:outline-none
  "
        />

        <div className="grid grid-cols-3 items-center text-xs text-[#9AA1AC]">
          {/* Left */}
          <span>{remaining} chars left</span>

          {/* Center */}
          <span
            className="text-center transition-opacity"
            style={{ opacity: props.powInfo ? 1 : 0 }}
          >
            {props.powInfo ?? 'placeholder'}
          </span>

          {/* Right */}
          <div className="justify-self-end">
            <Button
              type="submit"
              variant="outline"
              className="
    rounded-lg
    border
    border-[#9AA1AC]
    bg-[#2A323C]
    text-[#9AA1AC]
    hover:bg-[#3A414C]
    hover:text-[#9AA1AC]
    transition-colors
  "
              disabled={props.disabled || props.text.length === 0}
            >
              Submit
            </Button>
          </div>
        </div>
      </form>
    </div>
  )
}

function Logo(props: {
  open: boolean
  onToggle(): void
  count: number | null
}) {
  return (
    <div className="mt-12 mb-8 text-center px-4">
      <div
        onClick={props.onToggle}
        className="
          cursor-pointer
          font-mono
          text-[#8FD3E8]
          transition-all
          duration-200
          hover:text-[#B6E7F5]
          hover:tracking-wide
          active:scale-[0.99]
          select-none
          mx-auto
          break-words
        "
        title="Click to reveal more"
      >
        initialsDB. Post and search messages.
      </div>

      {/* Reserved space */}
      <div
        className="mt-3 text-sm text-[#8FD3E8] max-w-xl mx-auto transition-opacity duration-200"
        style={{
          minHeight: '1.5rem',
          opacity: props.open ? 1 : 0,
          pointerEvents: props.open ? 'auto' : 'none',
        }}
      >
        No accounts. No edits. Messages stay forever:{' '}
        {props.count !== null ? props.count.toLocaleString() : '…'}.{' '}
        <a
          className="text-[#9AA1AC] hover:text-[#F5F5F5]"
          href="https://www.buymeacoffee.com/aabbtree77"
          target="_blank"
          rel="noopener noreferrer"
        >
          Buy me a coffee…
        </a>
      </div>
    </div>
  )
}

// ==================================================
// App
// ==================================================

export default function App() {
  const [state, setState] = useState<AppState>({ tag: 'idle' })
  const [query, setQuery] = useState('')
  const [items, setItems] = useState<Listing[]>([])
  const [postText, setPostText] = useState('')
  const [powInfo, setPowInfo] = useState<string | null>(null)
  const [postOpen, setPostOpen] = useState(false)
  const [logoOpen, setLogoOpen] = useState(false)
  const [totalCount, setTotalCount] = useState<number | null>(null)

  // Hover status for the Hemingway hint
  const [hoverStatus, setHoverStatus] = useState<string | null>(null)

  useEffect(() => {
    const ac = new AbortController()
    fetchCount(ac.signal)
      .then(setTotalCount)
      .catch(() => {})
    return () => ac.abort()
  }, [])

  const [statusQueue, setStatusQueue] = useState<StatusMessage[]>([]) // removed initial Hemingway message

  const pushStatus = useCallback((text: string, type: StatusType) => {
    setStatusQueue((q) => [...q, { id: Date.now(), text, type }])
  }, [])

  const currentStatus = statusQueue.at(-1) ?? null

  const searchAbort = useRef<AbortController | null>(null)
  const postAbort = useRef<AbortController | null>(null)

  const PAGE_SIZE = 30
  const locked = state.tag === 'searching' || state.tag === 'pow'
  const isSearching = state.tag === 'searching'

  const startSearch = useCallback(async () => {
    if (locked) return

    setPostOpen(false)
    setState({ tag: 'idle' })

    if (!query.trim()) {
      setItems([])
      pushStatus('Search cleared.', 'info')
      return
    }

    searchAbort.current?.abort()
    searchAbort.current = new AbortController()

    setState({ tag: 'searching' })

    try {
      const res = await searchAPI(
        query,
        PAGE_SIZE,
        null,
        searchAbort.current.signal,
      )

      setItems(res.items)
      setState({ tag: 'search', cursor: res.next_cursor ?? null })
      pushStatus(`Results: ${res.items.length}`, 'info')
    } catch {
      setState({ tag: 'idle' })
      pushStatus('Search failed.', 'error')
    }
  }, [query, locked, pushStatus])

  const loadMore = useCallback(async () => {
    if (state.tag !== 'search' || !state.cursor || locked) return

    searchAbort.current?.abort()
    searchAbort.current = new AbortController()
    setState({ tag: 'searching' })

    try {
      const res = await searchAPI(
        query,
        PAGE_SIZE,
        state.cursor,
        searchAbort.current.signal,
      )

      setItems((prev) => [...prev, ...res.items])
      setState({ tag: 'search', cursor: res.next_cursor ?? null })
    } catch {
      setState({ tag: 'search', cursor: state.cursor })
      pushStatus('Search failed.', 'error')
    }
  }, [state, query, locked, pushStatus])

  const submitPost = async () => {
    if (state.tag !== 'posting') return

    if (postText.length > 255) {
      pushStatus('Post exceeds 255 characters.', 'error')
      return
    }

    postAbort.current?.abort()
    postAbort.current = new AbortController()

    try {
      setState({ tag: 'pow' })

      const pow = await getChallenge()

      const nonce = await solvePoW(
        pow.challenge,
        pow.difficulty,
        pow.ttl_secs,
        (tries, remaining) => {
          setPowInfo(`${tries.toLocaleString()} tries · ${remaining}s`)
        },
      )

      await postAPI(
        postText,
        { challenge: pow.challenge, nonce, token: pow.token },
        postAbort.current.signal,
      )
      setTotalCount((n) => (n == null ? n : n + 1))

      setPostText('')
      setPowInfo(null)
      setPostOpen(false)
      setState({ tag: 'idle' })
      pushStatus('Post saved.', 'info')
    } catch {
      setPowInfo(null)
      setState({ tag: 'posting' })
      pushStatus('PoW did not complete, submit again.', 'error')
    }
  }

  // Handlers for hover hint
  const handleMouseEnterPost = useCallback(() => {
    setHoverStatus(HOVER_MESSAGE)
  }, [])

  const handleMouseLeavePost = useCallback(() => {
    setHoverStatus(null)
  }, [])

  const handleTogglePost = useCallback(() => {
    if (locked) return
    setPostOpen((v) => !v)
    setState((s) =>
      s.tag === 'posting' ? { tag: 'idle' } : { tag: 'posting' },
    )
    setHoverStatus(null) // clear hint when toggled
  }, [locked])

  return (
    // Outer container: full screen, no horizontal scroll, flex column
    <div className="min-h-screen bg-[#2A323C] flex flex-col overflow-x-hidden">
      {/* Main content area with top padding to position control line at 20% from top */}
      <div className="flex-1 w-[90vw] md:w-[40vw] mx-auto pt-[5vh] lg:pt-[20vh]">
        <ControlLine
          query={query}
          disabled={locked}
          postOpen={postOpen}
          onQueryChange={setQuery}
          onSearch={startSearch}
          onTogglePost={handleTogglePost}
          onMouseEnterPost={handleMouseEnterPost}
          onMouseLeavePost={handleMouseLeavePost}
        />

        <StatusLine message={currentStatus} hoverMessage={hoverStatus} />

        {postOpen && (state.tag === 'posting' || state.tag === 'pow') && (
          <PostForm
            text={postText}
            disabled={state.tag === 'pow'}
            powInfo={powInfo}
            onTextChange={setPostText}
            onSubmit={submitPost}
          />
        )}

        {state.tag === 'search' && (
          <SearchResults
            items={items}
            loading={isSearching}
            hasMore={Boolean(state.cursor)}
            onLoadMore={loadMore}
          />
        )}
      </div>

      <div className="w-[90vw] md:w-[40vw] mx-auto mb-6">
        <Logo
          open={logoOpen}
          onToggle={() => setLogoOpen((v) => !v)}
          count={totalCount}
        />
      </div>
    </div>
  )
}
