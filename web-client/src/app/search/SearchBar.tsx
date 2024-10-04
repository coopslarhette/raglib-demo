'use client'

import React from 'react'
import styles from './SearchBar.module.css'
import { Button } from '@mui/base'

interface SearchBarProps {
    query: string
    setQuery: (query: string) => void
    onSearch: () => void
}

export default function SearchBar({
    query,
    setQuery,
    onSearch,
}: SearchBarProps) {
    const handleKeyPress = (event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
            onSearch()
        }
    }

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        onSearch()
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
