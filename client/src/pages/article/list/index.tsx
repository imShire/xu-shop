import { Image, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useInfiniteQuery } from '@tanstack/react-query'
import PullList from '@/components/PullList'
import { getArticles } from '@/services/article'
import type { Article } from '@/services/article'
import './index.scss'

const PAGE_SIZE = 20

export default function ArticleListPage() {
  const { data, isFetching, isFetchingNextPage, hasNextPage, fetchNextPage } = useInfiniteQuery({
    queryKey: ['articles', 'list'],
    queryFn: ({ pageParam = 1 }) =>
      getArticles({ page: pageParam as number, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const loaded = lastPage.page * lastPage.page_size
      return loaded >= lastPage.total ? undefined : lastPage.page + 1
    },
  })

  const allArticles: Article[] = data?.pages.flatMap((page) => page.list) ?? []
  const isFirstLoad = isFetching && allArticles.length === 0

  function handleLoadMore() {
    if (!isFetchingNextPage && hasNextPage) {
      void fetchNextPage()
    }
  }

  return (
    <View className='article-list-page'>
      <PullList
        data={allArticles}
        loading={isFirstLoad}
        hasMore={hasNextPage ?? false}
        onLoadMore={handleLoadMore}
        emptyTitle='暂无文章'
        emptyDescription='还没有发布任何文章'
        keyExtractor={(item) => item.id}
        renderItem={(article) => (
          <View
            className='article-list-page__item'
            onClick={() =>
              void Taro.navigateTo({ url: `/pages/article/detail/index?id=${article.id}` })
            }
          >
            {article.cover ? (
              <Image
                className='article-list-page__cover'
                src={article.cover}
                mode='aspectFill'
              />
            ) : null}
            <View className='article-list-page__info'>
              <Text className='article-list-page__title'>{article.title}</Text>
              {article.published_at ? (
                <Text className='article-list-page__date'>
                  {article.published_at.slice(0, 10)}
                </Text>
              ) : null}
            </View>
          </View>
        )}
      />
    </View>
  )
}
