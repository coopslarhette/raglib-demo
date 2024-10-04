'use client'

import React, { useState } from 'react'
import styles from './SearchBar.module.css'
import { Button } from '@mui/base'

interface SearchBarProps {
    initialQuery: string
    onSearch: (query: string) => void
}

export default function SearchBar({ initialQuery, onSearch }: SearchBarProps) {
    const [query, setQuery] = useState(initialQuery)
    const handleKeyPress = (event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
            onSearch(query)
        }
    }

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        onSearch(query)
    }

    return (
        <div className={styles.searchBar}>
            <form onSubmit={handleSubmit} className={styles.searchForm}>
                <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    onKeyDown={handleKeyPress}
                    placeholder="why are peppers spicy"
                    className={styles.searchInput}
                />
                <Button
                    type="submit"
                    disabled={query.length === 0}
                    className={styles.searchButton}
                >
                    Search
                </Button>
            </form>
        </div>
    )
}
