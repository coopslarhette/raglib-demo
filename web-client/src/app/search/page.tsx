import styles from './page.module.css'
import SearchLandingPage from '@/app/search/SearchLandingPage'

export default function Search() {
    return (
        <div className={styles.root}>
            <h2 className={styles.header}>RAGLib</h2>
            <SearchLandingPage />
        </div>
    )
}
