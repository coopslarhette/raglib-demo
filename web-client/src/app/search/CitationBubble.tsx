import React from 'react'
import styles from './CitationBubble.module.css'

type CitationProps = {
    citationNumber: number
    onClick?: () => void
    onMouseEnter: () => void
    onMouseLeave: () => void
}

export function CitationBubble({
    citationNumber,
    onClick,
    onMouseEnter,
    onMouseLeave,
}: CitationProps) {
    return (
        <span
            className={styles.citation}
            onClick={onClick}
            onMouseEnter={onMouseEnter}
            onMouseLeave={onMouseLeave}
        >
            {citationNumber}
        </span>
    )
}
