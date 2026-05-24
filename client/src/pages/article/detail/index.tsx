import { Image, RichText, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery } from '@tanstack/react-query'
import { Skeleton } from '@/ui/nutui'
import { getArticle } from '@/services/article'
import './index.scss'

export default function ArticleDetailPage() {
  const id = Taro.getCurrentInstance().router?.params?.id ?? ''

  const { data: article, isLoading, isError } = useQuery({
    queryKey: ['article', id],
    queryFn: () => getArticle(id),
    enabled: Boolean(id),
  })

  function handleBack() {
    const pages = Taro.getCurrentPages()
    if (pages.length > 1) {
      void Taro.navigateBack()
    } else {
      void Taro.switchTab({ url: '/pages/home/index' })
    }
  }

  if (isLoading) {
    return (
      <View className='article-detail-page article-detail-page--loading'>
        <Skeleton animated rows={8} />
      </View>
    )
  }

  if (isError || !article) {
    return (
      <View className='article-detail-page article-detail-page--error'>
        <Text className='article-detail-page__error-text'>文章不存在或已下线</Text>
        <View className='article-detail-page__back-btn' onClick={handleBack}>
          <Text className='article-detail-page__back-text'>返回</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='article-detail-page'>
      <View className='article-detail-page__content'>
        <Text className='article-detail-page__title'>{article.title}</Text>
        {article.published_at ? (
          <Text className='article-detail-page__date'>
            {article.published_at.slice(0, 10)}
          </Text>
        ) : null}
        {article.cover ? (
          <Image
            className='article-detail-page__cover'
            src={article.cover}
            mode='widthFix'
            style={{ width: '100%' }}
          />
        ) : null}
        {article.content ? (
          <View className='article-detail-page__body'>
            <RichText nodes={article.content} />
          </View>
        ) : null}
      </View>

      <View className='article-detail-page__footer'>
        <View className='article-detail-page__back-btn' onClick={handleBack}>
          <Text className='article-detail-page__back-text'>返回</Text>
        </View>
      </View>
    </View>
  )
}
