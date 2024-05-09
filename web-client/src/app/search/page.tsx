import styles from './page.module.css'
import SearchLandingPage from '@/app/search/SearchLandingPage'

export default function Search() {
    return <div className={styles.root}>
        <p>RAGLib</p>
        <SearchLandingPage />
    </div>
}
