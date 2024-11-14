import { AnswerChunk, ChunkType, SourceDocument } from '@/app/search/types'
import { useRouter } from 'next/navigation'
import { useCallback, useEffect, useReducer, useRef, useState } from 'react'
import { toSearchURL } from '@/api'

interface AnswerStreamState {
    documents: SourceDocument[]
    answerChunks: AnswerChunk[]
}

type AnswerStreamAction =
    | { type: 'ADD_ANSWER_CHUNK'; payload: AnswerChunk }
    | { type: 'SET_DOCUMENTS'; payload: SourceDocument[] }
    | { type: 'RESET' }

// Lol this might be over-engineered
function answerStreamReducer(
    state: AnswerStreamState,
    action: AnswerStreamAction
): AnswerStreamState {
    switch (action.type) {
        case 'ADD_ANSWER_CHUNK':
            return {
                ...state,
                answerChunks: [...state.answerChunks, action.payload],
            }
        case 'SET_DOCUMENTS':
            return { ...state, documents: action.payload }
        case 'RESET':
            return { documents: [], answerChunks: [] }
        default:
            return state
    }
}

export const useAnswerStream = (initialQuery: string) => {
    const router = useRouter()
    const [isResponseLoading, setIsResponseLoading] = useState(false)
    const [{ answerChunks, documents }, dispatch] = useReducer(
        answerStreamReducer,
        {
            documents: [],
            answerChunks: [],
        }
    )
    const eventSourceRef = useRef<EventSource | null>(null)

    // Do this so we show loading spinner when navigating to a results page
    // and response is still loading
    useEffect(() => {
        if (initialQuery.length > 0) {
            setIsResponseLoading(true)
        }
    }, [initialQuery])

    const eventHandler =
        (eventType: ChunkType, eventSource: EventSource) =>
        (event: MessageEvent) => {
            setIsResponseLoading(false)
            const data = JSON.parse(event.data)

            switch (eventType) {
                case 'text':
                case 'citation':
                case 'codeblock':
                    dispatch({
                        type: 'ADD_ANSWER_CHUNK',
                        payload: {
                            type: eventType,
                            value: data,
                            // Note: technically might be misusing lastEventId here according to Mozilla spec
                            ID: event.lastEventId,
                        },
                    })
                    break
                case 'documentsreference':
                    dispatch({ type: 'SET_DOCUMENTS', payload: data })
                    break
                case 'done':
                    eventSource.close()
                    break
            }
        }

    const handleSearch = useCallback(
        (query: string) => {
            router.push(`/search?q=${encodeURIComponent(query)}`, {
                scroll: false,
            })

            // Clean up existing event source if it exists, this prevents two sets of
            // stream event listeners writing to the answerChunks/documents state at once
            // This is a fix to the problem, however I'm not sure if it could be avoided entirely
            // if I was more of an expert about using SSE streams with React/had a different arch
            if (eventSourceRef.current !== null) {
                eventSourceRef.current.close()
                eventSourceRef.current = null
            }
            setIsResponseLoading(true)
            dispatch({ type: 'RESET' })

            let lastEventData = null
            const handleError = (err: Event) => {
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
                    console.log(
                        'No data was received before the error occurred'
                    )
                }
                eventSource.close()
                setIsResponseLoading(false)
            }

            const eventSource = new EventSource(toSearchURL(query), {
                withCredentials: true,
            })
            // Save so the new event source object so it can be properly closed when a new
            // handleSearch call comes in
            eventSourceRef.current = eventSource
            ;(
                [
                    'text',
                    'citation',
                    'documentsreference',
                    'codeblock',
                    'done',
                ] as ChunkType[]
            ).forEach((eventType) => {
                eventSource.addEventListener(
                    eventType,
                    eventHandler(eventType, eventSource)
                )
            })

            eventSource.onerror = handleError
        },
        [router]
    )

    return {
        handleSearch,
        isResponseLoading,
        answerChunks,
        documents,
    }
}
