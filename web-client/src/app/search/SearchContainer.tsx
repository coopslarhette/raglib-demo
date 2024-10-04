'use client'

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import dynamic from 'next/dynamic'
import { AnswerChunk, EventType, SourceDocument } from './types'
import SearchBar from './SearchBar'
import { toSearchURL } from '@/api'
import { CircularProgress } from '@mui/material'

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
    const [isResponseLoading, setIsResponseLoading] = useState(
        initialQuery.length > 0
    )

    const router = useRouter()

    useEffect(() => {
        if (initialQuery.length > 0) {
            handleSearch()
        }
    }, [initialQuery])

    const handleSearch = () => {
        router.push(`/search?q=${encodeURIComponent(query)}`, { scroll: false })
        setIsResponseLoading(true)
        setDocuments([])
        setAnswerChunks([])

        const withSpinnerCanceller = (
            listener: (this: EventSource, event: MessageEvent<any>) => any
        ) => {
            function wrapped(this: EventSource, event: MessageEvent<any>) {
                setIsResponseLoading(false)
                listener.call(this, event)
            }
            return wrapped
        }

        const eventSource = new EventSource(toSearchURL(query), {
            withCredentials: true,
        })

        // We save this to debug in onerror handler
        let lastEventData: string | null = null

        const eventHandler = (eventType: EventType) =>
            withSpinnerCanceller((event: MessageEvent) => {
                lastEventData = event.data
                const data = JSON.parse(event.data)
                switch (eventType) {
                    case 'text':
                        setAnswerChunks((prev) => [
                            ...prev,
                            {
                                ...data,
                                type: 'text',
                                value: data as string,
                            },
                        ])
                        break
                    case 'citation':
                        setAnswerChunks((prev) => [
                            ...prev,
                            {
                                ...data,
                                type: 'citation',
                                value: data as number,
                            },
                        ])
                        break
                    case 'codeblock':
                        setAnswerChunks((prev) => [
                            ...prev,
                            {
                                ...data,
                                type: 'codeblock',
                                value: data as string,
                            },
                        ])
                        break
                    case 'documentsreference':
                        setDocuments(data as SourceDocument[])
                        break
                    case 'done':
                        eventSource.close()
                        break
                }
            })

        ;(
            [
                'text',
                'citation',
                'documentsreference',
                'codeblock',
                'done',
            ] as EventType[]
        ).forEach((eventType) => {
            eventSource.addEventListener(eventType, eventHandler(eventType))
        })

        eventSource.onerror = (err) => {
            console.error('There was an error with the event source')

            if (lastEventData) {
                try {
                    const parsedData = JSON.parse(lastEventData)
                    console.log('Last received data:', parsedData)
                    if ('error' in parsedData) {
                        console.error('Error details:', parsedData.error)
                    }
                } catch (parseError) {
                    console.log(
                        'Unable to parse last received data:',
                        lastEventData
                    )
                }
            } else {
                console.log('No data was received before the error occurred')
            }

            // might want to implement a reconnection strategy here
            // For now, we'll just close the connection
            eventSource.close()
            setIsResponseLoading(false)
        }
    }

    return (
        <>
            <SearchBar
                query={query}
                setQuery={setQuery}
                onSearch={handleSearch}
            />
            {isResponseLoading ? (
                <CircularProgress style={{ color: 'var(--brand-teal)' }} />
            ) : (
                <SearchResults
                    documents={documents}
                    answerChunks={answerChunks}
                />
            )}
        </>
    )
}
