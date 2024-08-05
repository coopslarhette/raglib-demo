import React from 'react'
import styles from './SourceCard.module.css'
import {
    Card,
    CardActionArea,
    CardContent,
    CardMedia,
    Link,
    Typography,
} from '@mui/material'
import { WebReference } from '@/app/search/types'
import clsx from 'clsx'
import { dateToHumanReadable } from '@/utils'

type SourceCardProps = {
    source: WebReference
    isHoveredViaCitation: boolean
}

export function SourceCard({
    source: { title, link, displayedLink, date, author, favicon, thumbnail },
    isHoveredViaCitation,
}: SourceCardProps) {
    const handleClick = () => {
        window.open(link, '_blank', 'noopener noreferrer')
    }

    return (
        <Card className={styles.cardRoot}>
            <CardActionArea
                onClick={handleClick}
                className={clsx(styles.cardActionArea, {
                    [styles.hoveredViaCitation]: isHoveredViaCitation,
                })}
            >
                {thumbnail && (
                    <CardMedia
                        component="img"
                        height="140"
                        image={thumbnail}
                        alt="Thumbnail"
                    />
                )}
                <CardContent className={styles.cardContent}>
                    {favicon && (
                        <div className={styles.header}>
                            <img
                                src={favicon}
                                alt="Favicon"
                                className={styles.favicon}
                            />
                        </div>
                    )}
                    <Typography variant="subtitle1" component="div">
                        {title}
                    </Typography>
                    <div className={styles.metadata}>
                        {date && (
                            <Typography variant="body2">
                                {dateToHumanReadable(date)}
                            </Typography>
                        )}
                    </div>
                </CardContent>
            </CardActionArea>
        </Card>
    )
}
