import { Button } from '@mui/base'
import React from 'react'
import styles from './SuggestQueries.module.css'

type SuggestedQueriesProps = {
    onQueryClick: (q: string) => void
}

const TO_SUGGEST = [
    'golang receiver methods',
    'van leeuwen ice cream founding date',
    'typescript partial type',
    'mt tam best hikes'
]

export default function SuggestedQueries({
    onQueryClick,
}: SuggestedQueriesProps) {
    const handleClick = (
        e: React.MouseEvent<HTMLButtonElement, MouseEvent>,
        q: string
    ) => {
        e.stopPropagation()
        onQueryClick(q)
    }

    return (
        <div className={styles.container}>
            {TO_SUGGEST.map((query) => (
                <Button
                    onClick={(e) => handleClick(e, query)}
                    className={styles.button}
                    key={query}
                >
                    {query}
                </Button>
            ))}
        </div>
    )
}
