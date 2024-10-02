'use client'

import { useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import dynamic from 'next/dynamic'
import { SourceDocument, AnswerChunk } from './types'
import SearchBar from './SearchBar'
import { toSearchURL } from '@/api'

const SearchResults = dynamic(() => import('./SearchResults'), { ssr: false })

interface SearchContainerProps {
    initialQuery: string
}

export default function SearchContainer({
    initialQuery,
}: SearchContainerProps) {
    const [query, setQuery] = useState(initialQuery)
    const [documents, setDocuments] = useState<SourceDocument[]>([])
    const [answerChunks, setAnswerChunks] = useState<AnswerChunk[]>([])
    const [isSearchResponseLoading, setIsSearchResponseLoading] =
        useState(false)

    const router = useRouter()
    const searchParams = useSearchParams()

    useEffect(() => {
        if (initialQuery) {
            handleSearch()
        }
    }, [initialQuery])

    const handleSearch = async () => {
        setIsSearchResponseLoading(true)
        setDocuments([])
        setAnswerChunks([])

        const withSpinnerCanceller = (
            listener: (this: EventSource, event: MessageEvent<any>) => any
        ) => {
            function wrapped(this: EventSource, event: MessageEvent<any>) {
                setIsSearchResponseLoading(false)
                listener.call(this, event)
            }
            return wrapped
        }

        const eventSource = new EventSource(toSearchURL(query), {
            withCredentials: true,
        })

        eventSource.addEventListener(
            'text',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)
                setAnswerChunks((prev) => [
                    ...prev,
                    { type: 'text', value: data, ID: data.ID },
                ])
            })
        )

        eventSource.addEventListener(
            'citation',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)
                setAnswerChunks((prev) => [
                    ...prev,
                    { type: 'citation', value: data, ID: data.ID },
                ])
            })
        )

        eventSource.addEventListener(
            'documentsreference',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)
                setDocuments(data)
            })
        )

        eventSource.addEventListener(
            'codeblock',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)
                setAnswerChunks((prev) => [
                    ...prev,
                    { type: 'codeblock', value: data, ID: data.ID },
                ])
            })
        )

        eventSource.addEventListener(
            'done',
            withSpinnerCanceller((event) => {
                eventSource.close()
            })
        )

        eventSource.onerror = (err) => {
            console.error(err)
            eventSource.close()
            setIsSearchResponseLoading(false)
        }
        setIsSearchResponseLoading(false)

        // Update URL with the current query
        router.push(`/search?q=${encodeURIComponent(query)}`, { scroll: false })
    }

    return (
        <>
            <SearchBar
                query={query}
                setQuery={setQuery}
                onSearch={handleSearch}
            />
            <SearchResults
                documents={documents}
                answerChunks={answerChunks}
                isLoading={isSearchResponseLoading}
            />
        </>
    )
}
