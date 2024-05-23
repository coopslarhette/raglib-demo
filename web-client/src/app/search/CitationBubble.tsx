import React from 'react'
import styles from './CitationBubble.module.css'

type CitationProps = {
    label: number
    onClick?: () => void
    onMouseEnter: () => void
    onMouseLeave: () => void
}

export function CitationBubble({
    label,
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
            {label}
        </span>
    )
}
