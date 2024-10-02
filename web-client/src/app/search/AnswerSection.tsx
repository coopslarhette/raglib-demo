'use client'

import React from 'react'
import styles from './AnswerSection.module.css'
import { AnswerChunk, SourceDocument } from './types'
import { CitationBubble } from './CitationBubble'
import CodeBlock from '@/app/CodeBlock'

interface AnswerChunkProps {
    ac: AnswerChunk
    documents: SourceDocument[]
    setHoveredCitationIndex: React.Dispatch<React.SetStateAction<number | null>>
}

export default function AnswerSection({ ac, documents, setHoveredCitationIndex }: AnswerChunkProps) {
    const handleCitationClick = (citationIndex: number) => {
        const link = documents[citationIndex]?.webReference?.link
        if (!link) return
        window.open(link, '_blank', 'noopener noreferrer')
    }

    switch (ac.type) {
        case 'text':
            return <span className={styles.answerText}>{ac.value}</span>
        case 'citation':
            return (
                <CitationBubble
                    onClick={() => handleCitationClick(ac.value)}
                    label={ac.value + 1}
                    onMouseEnter={() => setHoveredCitationIndex(ac.value)}
                    onMouseLeave={() => setHoveredCitationIndex(null)}
                />
            )
        case 'codeblock':
            const lines = ac.value.split('\n')
            const language = lines[0].match(/```(\w+)/)?.[1] || ''
            const code = lines.slice(1, -1).join('\n')
            return <CodeBlock language={language} code={code} />
        default:
            return <span>Unsupported answer chunk</span>
    }
}