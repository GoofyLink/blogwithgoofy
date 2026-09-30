import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, top: 80 }
    return { top: 0 }
  },
  routes: [
    {
      path: '/',
      component: () => import('@/layouts/FrontLayout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/views/front/Home.vue') },
        { path: 'essays', name: 'essays', component: () => import('@/views/front/Essays.vue') },
        { path: 'article/:id', name: 'article', component: () => import('@/views/front/ArticleDetail.vue') },
        { path: 'category/:id', name: 'category', component: () => import('@/views/front/Essays.vue') },
        { path: 'tag/:id', name: 'tag', component: () => import('@/views/front/Essays.vue') },
        { path: 'archives', name: 'archives', component: () => import('@/views/front/Archives.vue') },
        { path: 'links', name: 'links', component: () => import('@/views/front/Links.vue') },
        { path: 'about', name: 'about', component: () => import('@/views/front/About.vue') },
        { path: 'p/:slug', name: 'page', component: () => import('@/views/front/About.vue') },
        { path: 'learn', name: 'learn', component: () => import('@/views/front/LearnIndex.vue') },
        { path: 'learn/book/:id', name: 'learnBook', component: () => import('@/views/front/LearnBook.vue') },
        { path: 'learn/chapter/:id', name: 'learnChapter', component: () => import('@/views/front/LearnChapter.vue') },
        { path: 'learn/words', name: 'learnWords', component: () => import('@/views/front/Words.vue') },
        { path: 'learn/words/review', name: 'wordsReview', component: () => import('@/views/front/WordsReview.vue') },
        { path: 'novel', name: 'novel', component: () => import('@/views/front/NovelIndex.vue') },
        { path: 'novel/book/:id', name: 'novelBook', component: () => import('@/views/front/NovelBook.vue') },
        { path: 'novel/chapter/:id', name: 'novelChapter', component: () => import('@/views/front/NovelChapter.vue') },
        { path: 'anime', name: 'anime', component: () => import('@/views/front/Anime.vue') },
        { path: 'art', name: 'art', component: () => import('@/views/front/Art.vue') },
        { path: 'ai', name: 'ai', component: () => import('@/views/front/AiHub.vue') },
        { path: 'game', name: 'game', component: () => import('@/views/front/Game.vue') },
      ],
    },
    {
      path: '/admin/login',
      name: 'login',
      component: () => import('@/views/admin/Login.vue'),
    },
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', name: 'dashboard', component: () => import('@/views/admin/Dashboard.vue') },
        { path: 'profile', name: 'adminProfile', component: () => import('@/views/admin/ProfileSetting.vue') },
        { path: 'animes', name: 'adminAnimes', component: () => import('@/views/admin/AnimeManage.vue') },
        { path: 'arts', name: 'adminArts', component: () => import('@/views/admin/ArtManage.vue') },
        { path: 'ai', name: 'adminAi', component: () => import('@/views/admin/AiManage.vue') },
        { path: 'games', name: 'adminGames', component: () => import('@/views/admin/GameManage.vue') },
        { path: 'apilogs', name: 'adminApiLogs', component: () => import('@/views/admin/ApiLogManage.vue') },
        { path: 'articles', name: 'adminArticles', component: () => import('@/views/admin/ArticleList.vue') },
        { path: 'articles/new', name: 'articleNew', component: () => import('@/views/admin/ArticleEdit.vue') },
        { path: 'articles/edit/:id', name: 'articleEdit', component: () => import('@/views/admin/ArticleEdit.vue') },
        { path: 'books', name: 'adminBooks', component: () => import('@/views/admin/BookManage.vue') },
        { path: 'books/:id/chapters', name: 'adminChapters', component: () => import('@/views/admin/ChapterManage.vue') },
        { path: 'categories', name: 'adminCategories', component: () => import('@/views/admin/CategoryManage.vue') },
        { path: 'tags', name: 'adminTags', component: () => import('@/views/admin/TagManage.vue') },
        { path: 'comments', name: 'adminComments', component: () => import('@/views/admin/CommentManage.vue') },
        { path: 'links', name: 'adminLinks', component: () => import('@/views/admin/LinkManage.vue') },
        { path: 'pages', name: 'adminPages', component: () => import('@/views/admin/PageManage.vue') },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.token) {
    return { path: '/admin/login', query: { redirect: to.fullPath } }
  }
})

export default router
