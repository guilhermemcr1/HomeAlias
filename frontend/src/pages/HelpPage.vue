<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
</script>

<template>
  <div>
    <PageHeader title="Ajuda" description="O essencial para manter o seu endereço no ar." />
    <div class="space-y-4 max-w-[70ch]">
      <section class="surface p-5">
        <h2 class="font-display text-lg font-bold">Como funciona</h2>
        <ol class="mt-2 list-decimal space-y-1 pl-5 text-base-content/85">
          <li>Em <strong>Conexões</strong>, cadastre um token da Cloudflare com permissão Zone · DNS · Edit.</li>
          <li>Em <strong>Hosts</strong>, crie o nome que vai acompanhar o seu IP.</li>
          <li>Em <strong>Tokens</strong>, gere um token e configure o cliente (roteador, Docker ou Windows).</li>
          <li>No <strong>Painel</strong>, acompanhe se o host está online.</li>
        </ol>
      </section>
      <section class="surface p-5">
        <h2 class="font-display text-lg font-bold">O que significam os estados</h2>
        <dl class="mt-2 space-y-1 text-base-content/85">
          <div><dt class="inline font-semibold">Online:</dt> <dd class="inline">o cliente falou com o HomeAlias há pouco (até 15 min, por padrão).</dd></div>
          <div><dt class="inline font-semibold">Atenção:</dt> <dd class="inline">sem contato há mais que o esperado (até 1 h).</dd></div>
          <div><dt class="inline font-semibold">Offline:</dt> <dd class="inline">sem contato além do limite. Confira o cliente e a internet de casa.</dd></div>
          <div><dt class="inline font-semibold">Aguardando:</dt> <dd class="inline">o host existe, mas nenhum cliente atualizou ainda.</dd></div>
        </dl>
      </section>
      <section class="surface p-5">
        <h2 class="font-display text-lg font-bold">CGNAT</h2>
        <p class="mt-2 text-base-content/85">
          O IP detectado é o IP público visto pelo Cloudflare Tunnel. Em redes com CGNAT esse IP pode ser compartilhado
          e o host pode não ser alcançável de fora; o HomeAlias ainda assim atualiza o DNS para o IP detectado.
          Compare o IP do painel com o da sua operadora.
        </p>
      </section>
    </div>
  </div>
</template>
