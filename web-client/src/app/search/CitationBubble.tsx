import React from 'react'
import styles from './CitationBubble.module.css'

type CitationProps = {
    citationNumber: number
    onClick?: () => void
    onMouseEnter: () => void
    onMouseLeave: () => void
}

/*
 Only render the citation number as referencing a 1-indexed list when it is displayed.
 Otherwise, it references a 0-indexed list throughout the system
*/
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
            {citationNumber + 1}
        </span>
    )
}
